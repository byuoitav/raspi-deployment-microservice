package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"

	"github.com/fatih/color"

	"golang.org/x/crypto/ssh"
)

type device struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

type elkReport struct {
	Hostname  string `json:"hostname"`
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
}

var sshConfig = &ssh.ClientConfig{
	User: os.Getenv("PI_SSH_USERNAME"),
	Auth: []ssh.AuthMethod{
		ssh.Password(os.Getenv("PI_SSH_PASSWORD")),
	},
}

var TIMER_DURATION = 3 * time.Minute

//deploys to all pi's with the given class and designation
//e.g. class = "av-control"
//e.g. desigation = "development"
func Deploy(class, designation string) error {

	log.Printf("%s", color.HiGreenString("[helpers] deployment started"))

	//scheduledDeployments[deploymentType] = false //why?? it seems like this code doesn't get executed if this line evaluates to true

	allDevices, err := GetAllDevices(designation)
	if err != nil {
		return err
	}

	environment, err := retrieveEnvironmentVariables(class, designation)
	if err != nil {
		return err
	}

	dockerCompose, err := RetrieveDockerCompose(class, designation)
	if err != nil {
		return errors.New(fmt.Sprintf("error fetching docker-compose file: %s", err.Error()))
	}

	for i := range allDevices {
		go SendCommand(allDevices[i].Address, environment, dockerCompose) // Start an update for each Pi
	}

	return nil
}

func DeployDevice(hostname string) (string, error) {

	log.Printf("[helpers] starting single deployment...")

	//hostname should be all caps - names in config DB are all caps
	allCaps := strings.ToUpper(hostname)

	//retrieve room from configuration database
	room, err := db.GetDB().GetRoom(hostname)
	if err != nil {
		msg := fmt.Sprintf("failed to get room: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	//build device name
	deviceName := strings.Split(allCaps, "-")[2]
	log.Printf("[helpers] looking for device: %s", deviceName)

	//get device class
	var deviceClass string
	for _, device := range room.Devices {

		log.Printf("[helpers] found device: %s of class: %s", device.Name, device.Type)

		if device.Name == deviceName { //found device

			deviceClass = device.Type.ID
		}
	}

	if len(deviceClass) == 0 { //if we don't find anything
		msg := "device class not found"
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	//get environment file based on the two IDs
	envFile, err := retrieveEnvironmentVariables(deviceClass, room.Designation)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error fetching environment variables: %s", err.Error()))
	}

	dockerCompose, err := RetrieveDockerCompose(deviceClass, room.Designation)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error fetching docker-compose file: %s", err.Error()))
	}

	dev, err := db.GetDB().GetDevice(hostname)
	if err != nil {
		log.Printf("error getting device")
		return "", err
	}

	go SendCommand(dev.Address, envFile, dockerCompose) // Start an update for the Pi

	log.Printf("deployment started")
	return "deployment started", nil
}

func GetAllDevices(deploymentType string) ([]structs.Device, error) {

	//allDevices goes to interface.go and uses the db interface to get all devices
	allDevices, err := db.GetDB().GetDevicesByRoleAndType("ControlProcessor", "Pi3")
	if err != nil {
		log.Printf("Error getting devices 3: %v", err.Error())
		return []structs.Device{}, err
	}

	log.Printf("All devices from database: %+v", allDevices)

	return allDevices, nil
}

func reportToELK(hostname string, err error) {
	log.Printf("Sending error to %s\n", os.Getenv("ELK_ADDRESS"))

	report := elkReport{Hostname: hostname, Timestamp: time.Now().Format(time.RFC3339), Action: "Deployment failed to start: " + err.Error()}
	data, err := json.Marshal(&report)
	if err != nil {
		log.Printf("Error sending error report for %s to %s", hostname, os.Getenv("ELK_ADDRESS"))
	}

	_, err = http.Post(os.Getenv("ELK_ADDRESS"), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error sending the following error report for %s to %s: %s.", hostname, os.Getenv("ELK_ADDRESS"), err.Error())
	}
}

func SendCommand(hostname, environment, docker string) error {
	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		log.Printf("Error dialing %s: %s", hostname, err.Error())
		reportToELK(hostname, err)
		return err
	}

	log.Printf("TCP connection established to %s", hostname)
	defer connection.Close()

	magicSession, err := connection.NewSession()
	if err != nil {
		log.Printf("%s", color.HiRedString("[helpers] error starting a session with %s: %s", hostname, err.Error()))
		reportToELK(hostname, err)
		return err
	}

	log.Printf("SSH session established with %s", hostname)

	longCommand := fmt.Sprintf("bash -c 'curl %s/%s --output /tmp/docker-compose-tmp.yml && curl %s/%s --output /home/pi/.environment-variables && curl %s/move-environment-variables.sh --output /home/pi/move-environment-variables.sh && chmod +x /home/pi/move-environment-variables.sh && /home/pi/move-environment-variables.sh && source /etc/environment && echo \"$(cat /tmp/docker-compose-tmp.yml)\" | envsubst > /tmp/docker-compose.yml && docker-compose -f /tmp/docker-compose.yml pull && docker stop $(docker ps -a -q) || true && docker rmi -f $(docker images -q --filter \"dangling=true\") || true && docker rm $(docker ps -a -q) || true && docker-compose -f /tmp/docker-compose.yml up -d' &> /tmp/deployment_logs.txt", os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"), docker, os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"), environment, os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"))

	log.Printf("Running the following command on %s: %s", hostname, longCommand)

	err = magicSession.Run(longCommand)
	if err != nil {
		log.Printf("%s", color.HiRedString("[helpers] error updating %s: %s", hostname, err.Error()))
		reportToELK(hostname, err)
		return err
	}

	log.Printf("%s", color.HiGreenString("[helpers] finished updating %s", hostname))

	return nil
}
