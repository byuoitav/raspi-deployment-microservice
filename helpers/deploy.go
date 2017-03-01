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

func Deploy() (string, error) {
	allDevices, err := GetDevices()
	if err != nil {
		return "", err
	}

	fileName, err := retrieveEnvironmentVariables()
	if err != nil {
		return "", err
	}

	for i := range allDevices {
		log.Printf("%+v", allDevices[i])

		err := SendCommand(allDevices[i].Address, fileName)
		if err != nil {
			log.Printf("Error updating %s at %s", allDevices[i].Name, allDevices[i].Address)

			log.Printf("Sending error to %s\n", os.Getenv("ELK_ADDRESS"))

			report := elkReport{Hostname: allDevices[i].Address, Timestamp: time.Now().Format(time.RFC3339), Action: "Deployment failed to start: " + err.Error()}
			data, err := json.Marshal(&report)

			if err != nil {
				log.Printf("Error sending error report.")
				continue
			}
			_, err = http.Post(os.Getenv("ELK_ADDRESS"), "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Error sending error report: %s.", err.Error())
			}
			continue
		}
	}

	log.Printf("Deployment finished.")
	return "Deployment started", nil
}

func GetDevices() ([]device, error) {
	client := &http.Client{}

	token, err := bearertoken.GetToken()
	if err != nil {
		return []device{}, err
	}

	req, _ := http.NewRequest("GET", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/devices/roles/ControlProcessor/types/pi", nil)
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

	log.Printf("%+v", allDevices)

	return allDevices, nil
}

func SendCommand(hostname string, fileName string) error {
	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		return err
	}

	log.Printf("TCP connection established.")
	defer connection.Close()

	magicSession, err := connection.NewSession()
	if err != nil {
		return err
	}

	defer magicSession.Close()
	log.Printf("SSH session established.")

	err = magicSession.Start("sh -c 'curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/docker-compose.yml --output /tmp/docker-compose.yml && curl " + os.Getenv("RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS") + "/" + fileName + " --output /home/pi/.environment-variables && source ~/.bashrc && docker-compose -f /tmp/docker-compose.yml up -d &'")
	if err != nil {
		return err
	}

	log.Print("Done.")

	return nil
}
