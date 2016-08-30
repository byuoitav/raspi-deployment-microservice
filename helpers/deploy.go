package helpers

import (
	"bytes"
	"io"
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

	devices, err := GetDevices()
	if err != nil {
		return "", err
	}

	for device := range devices {

	}

	return "Deployment started", nil
}

func GetDevices() ([]device, error) {
	response, err := http.Get("http://localhost:8006/devices/roles/ControlProcessor/types/pi")
	if err != nil {
		return []device{}, err
	}

	defer response.Body.Close()
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return []device{}, err
	}

	return []device{}, nil
}

func SendCommand(cmd, hostname string) (string, error) {
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

	err = session.Run("/usr/bin/scp deploy.sh")
	if err != nil {
		return "", err
	}

	session.Run(cmd)

	return hostname + ": " + stdoutBuf.String(), nil
}
