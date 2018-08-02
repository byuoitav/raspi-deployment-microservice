package helpers

import (
	"fmt"
	"io"

	"github.com/byuoitav/common/nerr"
	"golang.org/x/crypto/ssh"
)

// SSHAndRunCommand ssh's into an address, runs a command on it, and writes the output of that command to <output>
func SSHAndRunCommand(address, command string, output io.Writer) *nerr.E {
	// ssh into address
	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		return nerr.Translate(err).Addf("failed to ssh into %v", address)
	}
	defer client.Close()

	// open a new session with the client
	session, err := client.NewSession()
	if err != nil {
		return nerr.Translate(err).Addf("failed to open session with %v", address)
	}

	// get pipes
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get stdout pipe with %v", address)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get stderr pipe with %v", address)
	}

	// read from output pipes, write output to output
	go readWrite("stdout", stdout, output, 512*1)
	go readWrite("stderr", stderr, output, 512*1)

	err = session.Run("sudo systemctl enable contacts && sudo systemctl start contacts")
	if err != nil {
		return nerr.Translate(err).Addf("failed to run command on %v", address)
	}

	return nil
}

func readWrite(fromName string, from io.Reader, to io.Writer, bufSize int) {
	buffer := make([]byte, bufSize)
	for {
		n, err := from.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// write last few bytes
				to.Write(buffer[:n])
				to.Write([]byte(fmt.Sprintf("Finished reading from %s\n", fromName)))
				return
			}

			// write error to to
			to.Write([]byte(fmt.Sprintf("error reading from %s: %s\n", fromName, err)))
			return
		}

		// write bytes to to
		to.Write(buffer[:n])
	}
}
