package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/configuration-database-microservice/structs"
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

// deploys environment variables and docker containers to pi's
func Deploy(deploymentType string) error {

	log.Printf("%s", color.HiGreenString("[helpers] deployment started"))

	scheduledDeployments[deploymentType] = false //why?? it seems like this code doesn't get executed if this line evaluates to true

	allDevices, err := GetAllDevices(deploymentType)
	if err != nil {
		return err
	}

	fileName, err := retrieveEnvironmentVariables(deploymentType)
	if err != nil {
		return err
	}

	for i := range allDevices {
		go SendCommand(allDevices[i].Address, fileName, deploymentType) // Start an update for each Pi
	}

	return nil
}

func DeploySingle(hostname string) (string, error) {

	log.Printf("[helpers] starting single deployment...")

	room, err := GetRoom(hostname)
	if err != nil {
		msg := fmt.Sprintf("failed to get room: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	fileName, err := retrieveEnvironmentVariables(room.RoomDesignation)
	if err != nil {
		log.Printf("error getting env variables")
		return "", err
	}

	dev, err := GetDevice(hostname)
	if err != nil {
		log.Printf("error getting device")
		return "", err
	}

	go SendCommand(dev.Address, fileName, room.RoomDesignation) // Start an update for the Pi

	log.Printf("Deployment started")
	return "Deployment started", nil
}

func GetDevice(hostname string) (structs.Device, error) {

	log.Printf("Getting device information for %v", hostname)

	splitRoom := strings.Split(hostname, "-")
	if len(splitRoom) != 3 {
		msg := fmt.Sprintf("invalid hostname: %s", hostname)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return structs.Device{}, errors.New(msg)
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/buildings/%s/rooms/%s/devices/%s", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), splitRoom[0], splitRoom[1], splitRoom[2]), nil)

	err := SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set bearer token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return structs.Device{}, errors.New(msg)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Printf("response: %v", resp)
	if err != nil {
		log.Printf("Error getting device 1: %v", err.Error())
		return structs.Device{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	log.Printf("b: %s", b)
	if err != nil {
		log.Printf("Error getting device 2: %v", err.Error())
		return structs.Device{}, err
	}

	toReturn := structs.Device{}
	err = json.Unmarshal(b, &toReturn)
	if err != nil {
		log.Printf("Error getting device 3: %v", err.Error())
		return structs.Device{}, err
	}

	log.Printf("Device: %v", toReturn)

	return toReturn, nil
}

//TODO make this use the existing DBO package
func GetAllDevices(deploymentType string) ([]device, error) {
	client := &http.Client{}

	token, err := bearertoken.GetToken()
	if err != nil {
		return []device{}, err
	}
	log.Printf("Making request for all devices to: %v", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/deployment/devices/roles/ControlProcessor/types/pi/"+deploymentType)

	req, _ := http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/deployment/devices/roles/ControlProcessor/types/pi/"+deploymentType, nil)

	if deploymentType == "production" {
		req, _ = http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/devices/roles/ControlProcessor/types/pi", nil)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := client.Do(req)
	log.Printf("response: %v", resp)
	if err != nil {
		log.Printf("Error getting devices 1: %v", err.Error())
		return []device{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	log.Printf("b: %s", b)
	if err != nil {
		log.Printf("Error getting devices 2: %v", err.Error())
		return []device{}, err
	}

	allDevices := []device{}
	err = json.Unmarshal(b, &allDevices)
	if err != nil {
		log.Printf("Error getting devices 3: %v", err.Error())
		return []device{}, err
	}

	log.Printf("All devices from database: %+v", allDevices)

	return allDevices, nil
}

func GetRoom(hostname string) (structs.Room, error) {

	log.Printf("[helpers] getting room: %s", hostname)

	splitHostname := strings.Split(hostname, "-")
	if len(splitHostname) != 3 {
		msg := fmt.Sprintf("invalid hostname: %s", hostname)
		log.Printf("%s", color.HiRedString("[helplers] %s", msg))
		return structs.Room{}, errors.New(msg)
	}

	client := &http.Client{}

	req, _ := http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+splitHostname[0]+"/rooms/"+splitHostname[1], nil)

	err := SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("cannot set token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return structs.Room{}, errors.New(msg)
	}

	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to complete request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return structs.Room{}, errors.New(msg)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read body: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helplers] %s", msg))
		return structs.Room{}, errors.New(msg)
	}

	var room structs.Room
	err = json.Unmarshal(b, &room)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal struct: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helplers] %s", msg))
		return structs.Room{}, errors.New(msg)
	}

	//log.Printf("Device room from database: %+v", room)

	return room, nil
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

func SendCommand(hostname string, fileName string, deploymentType string) error {
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
		log.Printf("Error starting a session with %s: %s", hostname, err.Error())
		reportToELK(hostname, err)
		return err
	}

	log.Printf("SSH session established with %s", hostname)

	longCommand := "sh -c 'curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/docker-compose-" + deploymentType + ".yml --output /tmp/docker-compose.yml && curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/" + fileName + " --output /home/pi/.environment-variables && echo \"PI_HOSTNAME=$(cat /etc/hostname)\" >> /home/pi/.environment-variables && curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/move-environment-variables.sh --output /home/pi/move-environment-variables.sh && chmod +x /home/pi/move-environment-variables.sh && /home/pi/move-environment-variables.sh && . /etc/environment && docker-compose -f /tmp/docker-compose.yml pull && docker rmi $(docker images -q --filter \"dangling=true\") || true && docker stop $(docker ps -a -q) || true && docker rm $(docker ps -a -q) || true && docker-compose -f /tmp/docker-compose.yml up -d'"

	log.Printf("Running the following command on %s: %s", hostname, longCommand)

	err = magicSession.Run(longCommand)
	if err != nil {
		log.Printf("Error updating %s: %s", hostname, err.Error())
		reportToELK(hostname, err)
		return err
	}

	log.Printf("Finished updating %s", hostname)

	return nil
}
