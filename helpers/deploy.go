package helpers

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	l "github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/event-translator-microservice/elkreporting"

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
	Message   string `json:"msg"`
	Success   bool   `json:"success"`
}

var sshConfig = &ssh.ClientConfig{
	User: os.Getenv("PI_SSH_USERNAME"),
	Auth: []ssh.AuthMethod{
		ssh.Password(os.Getenv("PI_SSH_PASSWORD")),
	},
	Timeout: 5 * time.Second,
}

var TIMER_DURATION = 3 * time.Minute

func deployHelper(allDevices []structs.Device, class, designation string) ([]elkReport, error) {
	l.SetLevel("debug")

	var report []elkReport

	environment, err := retrieveEnvironmentVariables(class, designation)
	if err != nil {
		return report, err
	}

	dockerCompose, err := RetrieveDockerCompose(class, designation)
	if err != nil {
		return report, errors.New(fmt.Sprintf("error fetching docker-compose file: %s", err.Error()))
	}

	//make the channel to get the responses down
	respChan := make(chan elkReport, len(allDevices))
	l.L.Debugf("Sending %v comands", len(allDevices))
	for _, k := range allDevices {
		l.L.Debugf("sending to %v", k.ID)
	}

	for i := range allDevices {
		go SendCommand(allDevices[i].Address, environment, dockerCompose, respChan) // Start an update for each Pi
	}

	for i := 0; i < len(allDevices); i++ {
		cur := <-respChan

		l.L.Debugf("Got response from %v", cur.Hostname)
		l.L.Debugf("Waiting on %v more", len(allDevices)-(i+1))
		report = append(report, cur)
	}

	l.SetLevel("info")
	return report, nil

}

//deploys to all pi's with the given class and designation
//e.g. class = "av-control"
//e.g. desigation = "development"
func Deploy(class, designation string) ([]elkReport, error) {
	var report []elkReport

	l.L.Infof("%s", color.HiGreenString("[helpers] deployment started"))                                       //scheduledDeployments[deploymentType] = false //why?? it seems like this code doesn't get executed if this line evaluates to true
	allDevices, er := db.GetDB().GetDevicesByRoleAndTypeAndDesignation("ControlProcessor", class, designation) //TODO: Make the deployment process role-dependent.
	if er != nil {
		return report, er
	}

	return deployHelper(allDevices, class, designation)
}

func DeployBuilding(building, class, designation string) ([]elkReport, error) {
	var report []elkReport

	l.L.Infof("%s", color.HiGreenString("[helpers] deployment started"))                                       //scheduledDeployments[deploymentType] = false //why?? it seems like this code doesn't get executed if this line evaluates to true
	allDevices, er := db.GetDB().GetDevicesByRoleAndTypeAndDesignation("ControlProcessor", class, designation) //TODO: Make the deployment process role-dependent.
	if er != nil {
		return report, er
	}

	var toDeploy []structs.Device

	//filter by building
	for i := range allDevices {
		if strings.EqualFold(allDevices[i].ID[:strings.Index(allDevices[i].ID, "-")], building) {
			toDeploy = append(toDeploy, allDevices[i])
		}
	}

	return deployHelper(toDeploy, class, designation)
}

func DeployDevice(hostname string) (elkReport, error) {
	var report elkReport

	l.L.Infof("[helpers] starting single deployment...")

	hostname = strings.ToUpper(hostname)

	//retrieve room from configuration database
	room, err := db.GetDB().GetRoom(hostname[:strings.LastIndex(hostname, "-")])
	if err != nil {
		msg := fmt.Sprintf("failed to get room: %s", err.Error())
		l.L.Infof("%s", color.HiRedString("[helpers] %s", msg))
		return report, errors.New(msg)
	}

	l.L.Infof("[helpers] looking for device: %s", hostname)

	//get device class
	var deviceClass string
	for _, device := range room.Devices {

		l.L.Infof("[helpers] found device: %s of class: %s", device.Name, device.Type.ID)

		if device.ID == hostname { //found device

			deviceClass = device.Type.ID
		}
	}

	if len(deviceClass) == 0 { //if we don't find anything
		msg := "device class not found"
		l.L.Infof("%s", color.HiRedString("[helpers] %s", msg))
		return report, errors.New(msg)
	}

	//get environment file based on the two IDs
	envFile, err := retrieveEnvironmentVariables(deviceClass, room.Designation)
	if err != nil {
		return report, errors.New(fmt.Sprintf("error fetching environment variables: %s", err.Error()))
	}

	dockerCompose, err := RetrieveDockerCompose(deviceClass, room.Designation)
	if err != nil {
		return report, errors.New(fmt.Sprintf("error fetching docker-compose file: %s", err.Error()))
	}

	dev, err := db.GetDB().GetDevice(hostname)
	if err != nil {
		l.L.Infof("error getting device")
		return report, err
	}

	respChan := make(chan elkReport, 1)
	go SendCommand(dev.Address, envFile, dockerCompose, respChan) // Start an update for the Pi

	report = <-respChan

	l.L.Infof("deployment started")
	return report, nil
}

func reportToELK(hostname string, msg string, success bool) elkReport {

	var key string

	if success {
		key = "Successful"
	} else {
		key = "Failed"
	}

	report := elkReport{Hostname: hostname, Timestamp: time.Now().Format(time.RFC3339), Message: key + ": " + msg}

	splitName := strings.Split(hostname, "-")

	var e events.Event

	if len(splitName) != 3 {
		e = events.Event{
			Hostname:         hostname,
			Timestamp:        time.Now().Format(time.RFC3339),
			LocalEnvironment: false,
			Building:         "",
			Room:             "",
			Event: events.EventInfo{
				Type:           events.DEPLOYMENT,
				Requestor:      "",
				EventCause:     events.AUTOGENERATED,
				Device:         hostname,
				EventInfoKey:   key,
				EventInfoValue: msg,
			},
		}
	} else {
		e = events.Event{
			Hostname:         hostname,
			Timestamp:        time.Now().Format(time.RFC3339),
			LocalEnvironment: false,
			Building:         splitName[0],
			Room:             splitName[0] + "-" + splitName[1],
			Event: events.EventInfo{
				Type:           events.DEPLOYMENT,
				Requestor:      "",
				EventCause:     events.AUTOGENERATED,
				Device:         splitName[2][:strings.Index(splitName[2], ".")],
				EventInfoKey:   key,
				EventInfoValue: msg,
			},
		}
	}

	l.L.Debugf("Sending event to %v", os.Getenv("ELASTIC_API_EVENTS"))
	elkreporting.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS"), e, 3*time.Second)

	return report
}

func SendCommand(hostname, environment, docker string, respChan chan elkReport) error {
	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		msg := fmt.Sprintf("Error dialing %s: %s", hostname, err.Error())
		l.L.Infof(msg)

		val := reportToELK(hostname, msg, false)
		val.Success = false
		respChan <- val
		return err
	}

	l.L.Infof("ssh connection established to %s", hostname)
	defer connection.Close()

	magicSession, err := connection.NewSession()
	if err != nil {
		msg := fmt.Sprintf("error starting a session with %s: %s", hostname, err.Error())
		l.L.Infof(msg)

		val := reportToELK(hostname, msg, false)
		val.Success = false
		respChan <- val
		return err
	}

	//report that we started a connection and issued command

	respChan <- elkReport{
		Message:   "Connection started and command isssued.",
		Hostname:  hostname,
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.L.Infof("SSH session established with %s", hostname)

	longCommand := fmt.Sprintf("bash -c 'curl %s/%s --output /tmp/docker-compose-tmp.yml && curl %s/%s --output /home/pi/.environment-variables && curl %s/move-environment-variables.sh --output /home/pi/move-environment-variables.sh && chmod +x /home/pi/move-environment-variables.sh && /home/pi/move-environment-variables.sh && source /etc/environment && echo \"$(cat /tmp/docker-compose-tmp.yml)\" | envsubst > /tmp/docker-compose.yml && docker-compose -f /tmp/docker-compose.yml pull && docker stop $(docker ps -a -q) || true && docker rmi -f $(docker images -q --filter \"dangling=true\") || true && docker rm $(docker ps -a -q) || true && docker-compose -f /tmp/docker-compose.yml up -d' &> /tmp/deployment_logs.txt", os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"), docker, os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"), environment, os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS"))

	l.L.Infof("Running the following command on %s: %s", hostname, longCommand)

	err = magicSession.Run(longCommand)
	if err != nil {
		msg := fmt.Sprintf("%s", color.HiRedString("[helpers] error updating %s: %s", hostname, err.Error()))
		l.L.Infof(msg)
		reportToELK(hostname, msg, false)
		return errors.New(msg)
	}

	l.L.Infof("%s", color.HiGreenString("[helpers] finished updating %s", hostname))

	reportToELK(hostname, "Command successfully executed.", true)

	return nil
}
