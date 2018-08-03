package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"golang.org/x/crypto/ssh"
)

func NewSession(client *ssh.Client, output io.Writer) (*ssh.Session, *nerr.E) {
	// open a new session with the client
	session, err := client.NewSession()
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to open a new session")
	}

	log.L.Debugf("Successfully opened new session")

	// get output pipes
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to get stdout pipe")
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to get stderr pipe")
	}

	go readWrite("stdout", stdout, output, 512*1)
	go readWrite("stderr", stderr, output, 512*1)

	return session, nil
}

/*
// SSH ssh's into an address, executes the given function on the session, and writes all the output to the given <output>
func SSH(address string) *ssh.Client {
	// ssh into address
	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		return nerr.Translate(err).Addf("failed to ssh into %v", address)
	}
	defer client.Close()

	log.L.Debugf("Successfully connected to %v", address)

		log.L.Debugf("Successfully opened new session with %v", address)

		log.L.Debugf("Reading from stdout/stderr pipes (address %v)", address)

		// start a new shell with the address
		err = session.Shell()
		if err != nil {
			return nerr.Translate(err).Addf("failed to start shell with %v", address)
		}

		er := exec(stdin)
		if er != nil {
			return er.Addf("Failed to execute commands on session with %v", address)
		}

		stdin.Close()

		// wait for all commands to finish
		err = session.Wait()
		if err != nil {
			return nerr.Translate(err).Addf("something went wrong sending commands to %v", address)
		}

		return nil
}
*/

type file struct {
	Path        string
	Permissions os.FileMode
	Bytes       []byte
}

func scp(c *ssh.Client, output io.Writer, files ...file) *nerr.E {
	for _, file := range files {
		dir := path.Dir(file.Path)
		name := path.Base(file.Path)

		// open new session to write file in
		session, er := NewSession(c, output)
		if er != nil {
			return er.Addf("failed to open new session")
		}
		defer session.Close()

		log.L.Debugf("writing file %v to %v on %s", name, dir, c.RemoteAddr())

		// get stdin pipe
		stdin, err := session.StdinPipe()
		if err != nil {
			return nerr.Translate(err).Addf("failed to open stdin pipe with %s", c.RemoteAddr())
		}

		// run command
		err = session.Start("/usr/bin/scp -t " + dir)
		if err != nil {
			return nerr.Translate(err).Addf("failed to run scp command on %s", c.RemoteAddr())
		}

		fmt.Fprintf(stdin, "C%#o %d %s\n", file.Permissions, len(file.Bytes), name)
		io.Copy(stdin, bytes.NewReader(file.Bytes))
		fmt.Fprint(stdin, "\x00")
		stdin.Close()

		err = session.Wait()
		if err != nil {
			return nerr.Translate(err).Addf("something went wrong scp'ing %v to %s", file.Path, c.RemoteAddr())
		}

		log.L.Debugf("successfully sent file")
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
