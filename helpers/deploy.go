package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tmc/scp"

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

	for i := range allDevices {
		log.Printf("%+v", allDevices[i])

		err := SendCommand(allDevices[i].Address)
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
	response, err := http.Get(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/roles/ControlProcessor/types/pi")
	if err != nil {
		return []device{}, err
	}

	allDevices := []device{}
	err = json.NewDecoder(response.Body).Decode(&allDevices)
	if err != nil {
		return []device{}, err
	}

	return allDevices, nil
}

func SendCommand(hostname string) error {
	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		return err
	}
	log.Printf("TCP connection established.")
	defer connection.Close()

	sessionSCP, err := connection.NewSession()
	if err != nil {
		return err
	}
	log.Printf("SSH session established.")
	defer sessionSCP.Close()

	err = scp.CopyPath("update.sh", "/tmp", sessionSCP)
	if err != nil {
		return err
	}

	log.Printf("Copied update.sh to the /tmp directory.")

	sessionDeploy, err := connection.NewSession()
	if err != nil {
		return err
	}

	defer sessionDeploy.Close()

	err = sessionDeploy.Start(
		"export ELK_ADDRESS=" + os.Getenv("ELK_ADDRESS") +
			" && export CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS=" + os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") +
			" && /tmp/update.sh")
	if err != nil {
		return err
	}

	log.Print("Done.")

	return nil
}
