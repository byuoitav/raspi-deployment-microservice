package helpers

import (
	"log"

	"golang.org/x/crypto/ssh"
)

//@param active - true indicates monitoring the contact points, false indicates not monitoring the contact points
func UpdateContactState(hostname string, active bool) error {

	state := "Disabling"
	if active {
		state = "Enabling"
	}

	log.Printf("%s contact point monitoring on %s...", state, hostname)

	connection, err := ssh.Dial("tcp", hostname+":22", sshConfig)
	if err != nil {
		log.Printf("Error dialing %s: %s", hostname, err.Error())
		return err
	}

	log.Printf("TCP connection established to %s", hostname)
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		log.Printf("Error starting session with %s: %s", hostname, err.Error())
		return err
	}

	log.Printf("SSH session established with %s", hostname)

	if active {
		err = session.Run("sudo systemctl enable contacts && sudo systemctl start contacts")
		if err != nil {
			log.Printf("Error enabling contacts service: %s", err.Error())
			return err
		}

	} else {
		err = session.Run("sudo systemctl stop contacts")
		if err != nil {
			log.Printf("Error disabling contacts service: %s", err.Error())
			return err
		}
	}

	return nil

}
