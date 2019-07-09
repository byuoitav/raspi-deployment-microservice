package helpers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/event-translator-microservice/reporters"
	"github.com/byuoitav/raspi-deployment-microservice/socket"
	"golang.org/x/crypto/ssh"
)

const (
	dockerComposeFile = "/tmp/docker-compose.yml"
	envVarsFile       = "/tmp/environment"
)

var (
	sshConfig *ssh.ClientConfig
	services  = [...]string{
		"device-monitoring",
	}
)

// DeployReport is returned after attempting a deployment
type DeployReport struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"msg"`
	Success   bool   `json:"success"`
}

func init() {
	// get ssh key
	bucket := s3.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("AWS_BUCKET_REGION")),
	})

	resp, err := bucket.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("RASPI_DEPLOYMENT_S3_BUCKET")),
		Key:    aws.String(os.Getenv("AWS_DEPLOYMENT_KEY")),
	})
	if err != nil {
		log.L.Fatalf("failed to get aws deployment key: %v", err)
	}
	defer resp.Body.Close()
	log.L.Infof("Successfully got AWS deployment key.")

	// read key from response
	key, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Fatalf("unable to read private ssh key: %v", err)
	}

	// parse the pem encoded private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.L.Fatalf("unable to read parse private ssh key: %v", err)
	}

	// get pi username
	uname := os.Getenv("PI_SSH_USERNAME")
	if len(uname) == 0 {
		log.L.Fatalf("PI_SSH_USERNAME must be set.")
	}

	// build ssh config
	sshConfig = &ssh.ClientConfig{
		User: uname,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO should we check the host key..?
		Timeout:         5 * time.Second,
	}
}

// DeployByHostname deploys to one specific device with the corresponding hostname
func DeployByHostname(hostname string) (DeployReport, *nerr.E) {
	var report DeployReport
	log.L.Infof("Starting DeployByHostname to %v", hostname)

	// get room from database
	room, err := db.GetDB().GetRoom(hostname[:strings.LastIndex(hostname, "-")])
	if err != nil {
		return report, nerr.Translate(err).Addf("failed to get room for hostname %v", hostname)
	}

	log.L.Debugf("Got room %v, looking for hostname %v", room.ID, hostname)

	// find the specific device
	var device structs.Device
	for i := range room.Devices {
		log.L.Debugf("Checking %v...", room.Devices[i].ID)
		if strings.EqualFold(room.Devices[i].ID, hostname) {
			device = room.Devices[i]
			break
		}
	}

	// if the device wasn't found
	if len(device.Type.ID) == 0 {
		return report, nerr.Create(fmt.Sprintf("failed to find device %v", hostname), reflect.TypeOf("").String())
	}

	log.L.Debugf("Got device %v", device.ID)

	report, er := DeployToDevice(device, device.Type.ID, room.Designation)
	if er != nil {
		return report, er.Addf("failed to deploy to device %v", device.ID)
	}

	return report, nil
}

// DeployToDevice takes a slice of devices and gets all the data it needs to deploy
func DeployToDevice(device structs.Device, deviceType, designation string) (DeployReport, *nerr.E) {
	var report DeployReport
	var toScp []file
	var servicesToDeploy []string

	// get docker compose file
	log.L.Debugf("In DeployToDevice")
	dockerCompose, err := RetrieveDockerCompose(deviceType, designation)
	if err != nil {
		return report, nerr.Translate(err).Addf("unable to retrieve docker-compose file: %v", err)
	}
	//Replace $SYSTEM_ID with the actual id
	tmp := fmt.Sprintf("%s", dockerCompose)
	tmp = strings.ReplaceAll(tmp, "$SYSTEM_ID", device.ID)
	var byter bytes.Buffer
	byter.WriteString(tmp)
	dockerCompose = byter.Bytes()

	//Add the docker compose to the scp files
	toScp = append(toScp, file{
		Path:        dockerComposeFile + ".tmp",
		Permissions: 0644,
		Bytes:       dockerCompose,
	})

	// get files for services
	for _, service := range services {
		//files, serviceFileExists, err := GetServiceFromS3(service, designation)
		files, serviceFileExists, err := GetServiceFromCouch(service, designation, deviceType, device.ID)
		if err != nil {
			return report, nerr.Translate(err).Addf("unable to get service %v", service)
		}

		if serviceFileExists {
			servicesToDeploy = append(servicesToDeploy, service)
		}

		toScp = append(toScp, files...)
	}

	report = Deploy(device.Address, socket.Writer(device.Address), servicesToDeploy, toScp...)

	return report, nil
}

// Deploy deploys to a single pi
func Deploy(address string, output io.Writer, servicesToDeploy []string, files ...file) DeployReport {
	report := DeployReport{
		Address:   address,
		Timestamp: time.Now().Format(time.RFC3339),
		Success:   false,
	}

	if len(address) == 0 {
		report.Message = fmt.Sprintf("Address to deploy cannot be empty")
		return report
	}

	log.L.Infof("Deploying to %s. Deploying services: %s", address, servicesToDeploy)

	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		report.Message = fmt.Sprintf("failed to open connection with %v: %v", address, err)
		return report
	}

	log.L.Debugf("Successfully connected to %v", address)

	go func() {
		defer client.Close()

		session, er := NewSession(client, output)
		if er != nil {
			msg := fmt.Sprintf("unable to open new session with %v: %v", address, er.String())
			log.L.Warnf(msg)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		stdin, err := session.StdinPipe()
		if err != nil {
			msg := fmt.Sprintf("unable to open stdin pipe on %v: %v", address, err)
			log.L.Warnf(msg)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		err = session.Shell()
		if err != nil {
			msg := fmt.Sprintf("unable to start shell on %v: %v", address, err)
			log.L.Warnf(msg)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		log.L.Debugf("Started new shell on %s", address)

		// execute everything as root
		fmt.Fprintf(stdin, `/usr/bin/sudo su`+"\n")

		// write all script execution to a file
		fmt.Fprintf(stdin, `script -f /tmp/deployment.log`+"\n")

		// kill all services
		for _, service := range servicesToDeploy {
			fmt.Fprintf(stdin, `systemctl stop %s.service`+"\n", service)
		}

		// scp files over
		er = scp(client, output, files...)
		if er != nil {
			msg := fmt.Sprintf("failed to scp files to %v: %v", address, er.String())
			log.L.Warnf(msg)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		log.L.Infof("Successfully scp'd files to %s", address)

		// set up env vars
		fmt.Fprintf(stdin, `cp -f %s /etc/environment`+"\n", envVarsFile)
		fmt.Fprintf(stdin, `source /etc/environment`+"\n")

		// docker stuff
		fmt.Fprintf(stdin, `docker-compose -f %s pull`+"\n", dockerComposeFile)
		fmt.Fprintf(stdin, `docker stop $(docker ps -aq)`+"\n")
		fmt.Fprintf(stdin, `docker rm $(docker ps -aq)`+"\n")
		fmt.Fprintf(stdin, `docker-compose -f %s up -d`+"\n", dockerComposeFile)
		fmt.Fprintf(stdin, `docker system prune -af`+"\n")

		// services
		for _, service := range servicesToDeploy {
			serviceFile := fmt.Sprintf(`%s.service`, service)
			serviceFilePath := fmt.Sprintf(`/byu/%s/%s`, service, serviceFile)
			// systemdFilePath := fmt.Sprintf(`/lib/systemd/system/%s`, serviceFile)

			// fmt.Fprintf(stdin, `cp -f %s %s`+"\n", serviceFilePath, systemdFilePath)
			fmt.Fprintf(stdin, `systemctl daemon-reload`+"\n")
			fmt.Fprintf(stdin, `systemctl enable %s`+"\n", serviceFilePath)
			fmt.Fprintf(stdin, `systemctl stop %s`+"\n", serviceFile)
			fmt.Fprintf(stdin, `systemctl start %s`+"\n", serviceFile)
		}

		// stop printing script output
		fmt.Fprintln(stdin, `exit`)
		stdin.Close()

		// wait for all commands to execute
		err = session.Wait()
		if err != nil {
			msg := fmt.Sprintf("failed to deploy to %v: %v", address, err)
			log.L.Warnf(msg)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		// send success to elk
		msg := fmt.Sprintf("Successfully deployed to %v", address)
		log.L.Infof(msg)
		fmt.Fprintf(output, msg)
		reportToELK(address, msg, true)
	}()

	report.Success = true
	return report
}

func reportToELK(address, msg string, success bool) DeployReport {

	var key string
	if success {
		key = "Successful"
	} else {
		key = "Failed"
	}

	report := DeployReport{
		Address:   address,
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
		Success:   success,
	}

	hostname, _ := os.Hostname()

	e := events.Event{
		GeneratingSystem: hostname,
		Timestamp:        time.Now(),
		EventTags: []string{
			events.AutoGenerated,
		},
		// TODO get the actual deviceID
		TargetDevice: events.GenerateBasicDeviceInfo(address),
		AffectedRoom: events.GenerateBasicRoomInfo(address),
		Key:          key,
		Value:        msg,
		User:         "",
		Data:         nil,
	}
	/*
		e := events.Event{
			Hostname:         address,
			Timestamp:        time.Now().Format(time.RFC3339),
			LocalEnvironment: false,
			Building:         "",
			Room:             "",
			Event: events.EventInfo{
				Type:           events.DEPLOYMENT,
				Requestor:      "",
				EventCause:     events.AUTOGENERATED,
				Device:         address,
				EventInfoKey:   key,
				EventInfoValue: msg,
			},
		}

		splitName := strings.Split(address, "-")

		if len(splitName) == 3 {
			e.Building = splitName[0]
			e.Room = splitName[0] + "-" + splitName[1]
			e.Event.Device = splitName[2][:strings.Index(splitName[2], ".")]
		}
	*/

	log.L.Debugf("Sending event to %v", os.Getenv("ELASTIC_API_EVENTS"))
	reporters.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS"), e, 3*time.Second)

	return report
}
