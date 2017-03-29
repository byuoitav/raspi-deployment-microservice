package helpers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/authmiddleware/bearertoken"

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

func Deploy(deploymentType string) (string, error) {
	allDevices, err := GetDevices(deploymentType)
	if err != nil {
		return "", err
	}

	fileName, err := retrieveEnvironmentVariables()
	if err != nil {
		return "", err
	}

	for i := range allDevices {
		go SendCommand(allDevices[i].Address, fileName) // Start an update for each Pi
	}

	log.Printf("Deployment started")
	return "Deployment started", nil
}

func GetDevices(deploymentType string) ([]device, error) {
	client := &http.Client{}

	token, err := bearertoken.GetToken()
	if err != nil {
		return []device{}, err
	}

	req, _ := http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/"+deploymentType+"/devices/roles/ControlProcessor/types/pi", nil)

	if deploymentType == "production" {
		req, _ = http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/devices/roles/ControlProcessor/types/pi", nil)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := client.Do(req)
	if err != nil {
		return []device{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []device{}, err
	}

	allDevices := []device{}
	err = json.Unmarshal(b, &allDevices)
	if err != nil {
		return []device{}, err
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

func SendCommand(hostname string, fileName string) error {
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

	//defer magicSession.Close()
	log.Printf("SSH session established with %s", hostname)

	longCommand := "sh -c 'curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/docker-compose.yml --output /tmp/docker-compose.yml && curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/" + fileName + " --output /home/pi/.environment-variables && echo \"PI_HOSTNAME=$(cat /etc/hostname)\" >> /home/pi/.environment-variables && curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/move-environment-variables.sh --output /home/pi/move-environment-variables.sh && chmod +x /home/pi/move-environment-variables.sh && /home/pi/move-environment-variables.sh && . /etc/environment && docker-compose -f /tmp/docker-compose.yml pull && docker rmi $(docker images -q --filter \"dangling=true\") || true && docker stop $(docker ps -a -q) || true && docker rm $(docker ps -a -q) || true && docker-compose -f /tmp/docker-compose.yml up -d'"

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
