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

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"

	"golang.org/x/crypto/ssh"
)

type device struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	Type          string `json:"type"`
	Room          room   `json:"room"`
	DockerCompose string //name of docker compose file for the device
	Environment   string //name of environment file for device
}

type room struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json: "description"`
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

//deploys to all pi's on the given branch with the given role
func DeployDesignation(designation, role string) error {
	return nil
}

func DeployRoom(roomName, roleName string) error {

	log.Printf("[helpers] deploying to %s", color.HiGreenString(roomName))

	info := strings.Split(roomName, "-") //	splitting room name on hyphen yields building and room
	if len(info) < 2 {
		msg := fmt.Sprintf("invalid room name: %s", roomName)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return errors.New(msg)
	}

	log.Printf("[helpers] getting room...")
	room, err := dbo.GetRoomByInfo(info[0], info[1]) //	get room designation
	if err != nil {
		msg := fmt.Sprintf("room %s not found: %s", roomName, err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return errors.New(msg)
	}

	//	identify targets based on roleName
	log.Printf("[helpers] getting role ID...")
	roleId, err := GetRoleId(roleName)
	if err != nil {
		return err
	}

	log.Printf("[helpers] identified role ID: %s", color.HiGreenString("%d", roleId))

	log.Printf("[helpers] finding targets...")
	targets, err := dbo.GetDevicesByRoomIdAndRoleId(room.ID, int(roleId))
	if err != nil {
		msg := fmt.Sprintf("targets not found: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return nil
	}

	for _, target := range targets {

		go DeployDevice(target)
	}

	return nil
}

func DeployDevice(target structs.Device) error {

	log.Printf("[helpers] deploying to target: %s", color.HiGreenString(target.Name))

	docker, err := GetDeviceDocker(target)
	if err != nil {
		return err
	}

	log.Printf("%s", color.HiGreenString(docker))

	env, err := GetDeviceEnvironment(target)
	if err != nil {
		return err
	}

	log.Printf("%s", color.HiGreenString(env))

	return SendCommand(target.Address, env, docker) // Start an update for the Pi
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
func GetAllDevices(deviceRole, deploymentType string) (*map[int][]device, error) {

	url := fmt.Sprintf("%s/deployment/devices/roles/ControlProcessor/types/pi/%s", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), deploymentType)
	log.Printf("[helpers] making request for all devices to: %s", url)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	SetToken(req)

	resp, err := client.Do(req)
	log.Printf("response: %v", resp)
	if err != nil {
		msg := fmt.Sprintf("error making request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return &map[int][]device{}, errors.New(msg)
	}

	b, err := ioutil.ReadAll(resp.Body)
	log.Printf("b: %s", b)
	if err != nil {
		msg := fmt.Sprintf("error reading repsonse: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return &map[int][]device{}, errors.New(msg)
	}

	var devices []device
	err = json.Unmarshal(b, &devices)
	if err != nil {
		msg := fmt.Sprintf("error unmarshalling structs: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return &map[int][]device{}, errors.New(msg)
	}

	deviceMap := make(map[int][]device) //maps room IDs to devices

	for _, device := range devices {

		deviceMap[device.Room.Id] = append(deviceMap[device.Room.Id], device)
	}

	return &deviceMap, nil
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

	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + splitHostname[0] + "/rooms/" + splitHostname[1]

	log.Printf("[helpers] making request against url: %s", url)

	req, _ := http.NewRequest("GET", url, nil)

	err := SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set token: %s", err.Error())
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

	longCommand := fmt.Sprintf("bash -c 'curl %s/%s --output /tmp/docker-compose-tmp.yml && curl %s/%s --output /home/pi/.environment-variables && curl %s/move-environment-variables.sh --output /home/pi/move-environment-variables.sh && chmod +x /home/pi/move-environment-variables.sh && /home/pi/move-environment-variables.sh && source /etc/environment && echo \"$(cat /tmp/docker-compose-tmp.yml)\" | envsubst > /tmp/docker-compose.yml && docker-compose -f /tmp/docker-compose.yml pull && docker stop $(docker ps -a -q) || true && docker rmi -f $(docker images -q --filter \"dangling=true\") || true && docker rm $(docker ps -a -q) || true && docker-compose -f /tmp/docker-compose.yml up -d' &> /tmp/deployment_logs.txt", os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"), docker, os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"), environment, os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"))

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
