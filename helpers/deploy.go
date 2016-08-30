package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/ssh"
)

type device struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

var sshConfig = &ssh.ClientConfig{
	User: os.Getenv("PI_SSH_USERNAME"),
	Auth: []ssh.AuthMethod{
		ssh.Password(os.Getenv("PI_SSH_PASSWORD")),
	},
}

func Deploy(repo string) (string, error) {
	// "repo" will eventually allow for running just one update, but for now we're running updates on all the containers

	allDevices, err := GetDevices()
	if err != nil {
		return "", err
	}

	for i := range allDevices {
		log.Printf("%+v", allDevices[i])

		response, err := SendCommand(allDevices[i].Address)
		if err != nil {
			return "", err
		}

		log.Println(response)
	}

	return "Deployment started", nil
}

func GetDevices() ([]device, error) {
	response, err := http.Get("http://localhost:8006/devices/roles/ControlProcessor/types/pi")
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

func SendCommand(hostname string) (string, error) {
	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		return "", err
	}

	session, err := connection.NewSession()
	if err != nil {
		return "", err
	}

	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err = session.Run("cd /home/aveng")
	if err != nil {
		return "", err
	}

	err = session.Run("/usr/bin/scp -t ./deploy.sh")
	if err != nil {
		return "", err
	}

	session.Run("./deploy.sh")

	return hostname + ": " + stdoutBuf.String(), nil
}
