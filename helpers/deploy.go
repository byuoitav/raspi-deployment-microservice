package helpers

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/raspi-deployment-microservice/socket"
	"golang.org/x/crypto/ssh"
)

const (
	dockerComposeFile = "/tmp/docker-compose.yml"
	envVarsFile       = "/tmp/environment"
)

// DeployReport is returned after attempting a deployment
type DeployReport struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"msg"`
	Success   bool   `json:"success"`
}

var sshConfig *ssh.ClientConfig

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
		log.L.Fatalf("failed to get aws deployment key")
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

// DeployByHostname deploys to one specific device with the corrosponding hostname
func DeployByHostname(hostname string) ([]DeployReport, *nerr.E) {
	var reports []DeployReport
	log.L.Infof("Starting DeployByHostname to %v", hostname)

	// get room from database
	room, err := db.GetDB().GetRoom(hostname[:strings.LastIndex(hostname, "-")])
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to get room for hostname %v", hostname)
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
		return reports, nerr.Create(fmt.Sprintf("failed to find device %v", hostname), reflect.TypeOf("").String())
	}

	log.L.Debugf("Got device %v", device.ID)

	reports, er := DeployToDevices([]structs.Device{device}, device.Type.ID, room.Designation)
	if err != nil {
		return reports, er.Addf("failed to deploy to device %v", device.ID)
	}

	return reports, nil
}

// DeployByTypeAndDesignation deploys to all devices of the given type and designation
func DeployByTypeAndDesignation(deviceType, designation string) ([]DeployReport, *nerr.E) {
	var reports []DeployReport
	log.L.Infof("Deploying by type %v and designation %v", deviceType, designation)

	// TODO create type/designation function in db
	allDevices, err := db.GetDB().GetDevicesByRoleAndTypeAndDesignation("ControlProcessor", deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to get devices by type %v and designation %v", deviceType, designation)
	}

	log.L.Debugf("Got %v devices matching type %v and designation %v", len(allDevices), deviceType, designation)

	reports, er := DeployToDevices(allDevices, deviceType, designation)
	if err != nil {
		return reports, er.Addf("failed to deploy to devices by type %v and designation %v", deviceType, designation)
	}

	return reports, nil
}

// DeployByBuildingAndTypeAndDesignation deploys to all devices within the given building of the given type/designation
func DeployByBuildingAndTypeAndDesignation(building, deviceType, designation string) ([]DeployReport, *nerr.E) {
	var reports []DeployReport
	log.L.Infof("Deploying to building %v, with device type %v and designation %v", building, deviceType, designation)

	// TODO create type/designation function in db
	allDevices, err := db.GetDB().GetDevicesByRoleAndTypeAndDesignation("ControlProcessor", deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to get devices by type %v and designation %v", deviceType, designation)
	}

	log.L.Debugf("Filtering for devices in building %s", building)

	// filter out by building
	var buildingDevices []structs.Device
	for i := range allDevices {
		if strings.EqualFold(allDevices[i].ID[:strings.Index(allDevices[i].ID, "-")], building) {
			buildingDevices = append(buildingDevices, allDevices[i])
		}
	}

	log.L.Debugf("Got %v devices in building %v, with type %v and designation %v", len(buildingDevices), building, deviceType, designation)

	reports, er := DeployToDevices(allDevices, deviceType, designation)
	if err != nil {
		return reports, er.Addf("failed to deploy to devices in building %v by type %v and designation %v", building, deviceType, designation)
	}

	return reports, nil
}

// DeployToDevices takes a slice of devices and gets all the data it needs to deploy
func DeployToDevices(devices []structs.Device, deviceType, designation string) ([]DeployReport, *nerr.E) {
	var reports []DeployReport
	var reportsMu sync.Mutex

	// get env vars
	envVars, err := retrieveEnvironmentVariables(deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("unable to retrieve environment variables: %v", err)
	}

	// get docker compose file
	dockerCompose, err := RetrieveDockerCompose(deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("unable to retrieve docker-compose file: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(devices))

	// deploy to each device
	for i := range devices {
		go func(idx int) {
			report := Deploy(devices[idx].Address, envVars, dockerCompose, socket.Writer(devices[idx].Address))

			reportsMu.Lock()
			reports = append(reports, report)
			reportsMu.Unlock()

			wg.Done()
		}(i)
	}

	wg.Wait()
	return reports, nil
}

// Deploy deploys to a single pi
func Deploy(address string, envVars, dockerCompose []byte, output io.Writer) DeployReport {
	report := DeployReport{
		Address:   address,
		Timestamp: time.Now().Format(time.RFC3339),
		Success:   false,
	}

	if len(address) == 0 {
		report.Message = fmt.Sprintf("Address to deploy cannot be empty")
		return report
	}

	log.L.Infof("Deploying to %s", address)

	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		report.Message = fmt.Sprintf("failed to open connection with %v: %v", address, err)
		return report
	}

	log.L.Debugf("Successfully connected to %v", address)

	go func() {
		defer client.Close()
		// scp files over
		files := []file{
			file{
				Path:        envVarsFile,
				Permissions: 0644,
				Bytes:       envVars,
			},
			file{
				Path:        dockerComposeFile + ".tmp",
				Permissions: 0644,
				Bytes:       dockerCompose,
			},
		}
		er := scp(client, output, files...)
		if er != nil {
			msg := fmt.Sprintf("failed to scp files to %v: %v", address, er.String())
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		log.L.Debugf("Successfully scp'd files to %s", address)

		session, er := NewSession(client, output)
		if er != nil {
			msg := fmt.Sprintf("unable to open new session with %v: %v", address, er.String())
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		stdin, err := session.StdinPipe()
		if err != nil {
			msg := fmt.Sprintf("unable to open stdin pipe on %v: %v", address, err)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		err = session.Shell()
		if err != nil {
			msg := fmt.Sprintf("unable to start shell on %v: %v", address, err)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		log.L.Debugf("Started new shell on %s", address)

		// write all script execution to a file
		fmt.Fprintf(stdin, `script -f /tmp/deployment.log`+"\n")

		// set up env vars
		fmt.Fprintf(stdin, `echo "export PI_HOSTNAME=\"$(hostname)\"" >> %s`+"\n", envVarsFile)
		fmt.Fprintf(stdin, `sudo cp %s /etc/environment`+"\n", envVarsFile)
		fmt.Fprintf(stdin, `source /etc/environment`+"\n")
		fmt.Fprintf(stdin, `cat "%s" | envsubst > %s`+"\n", dockerComposeFile+".tmp", dockerComposeFile)

		// docker stuff
		fmt.Fprintf(stdin, `docker-compose -f %s pull`+"\n", dockerComposeFile)
		fmt.Fprintf(stdin, `docker stop $(docker ps -aq)`+"\n")
		fmt.Fprintf(stdin, `docker-compose -f %s up -d`+"\n", dockerComposeFile)
		fmt.Fprintf(stdin, `docker system prune -af`+"\n")

		// stop printing script output
		fmt.Fprintln(stdin, `exit`)
		stdin.Close()

		// wait for all commands to execute
		err = session.Wait()
		if err != nil {
			msg := fmt.Sprintf("failed to deploy to %v: %v", address, err)
			fmt.Fprintf(output, msg)
			reportToELK(address, msg, false)
			return
		}

		// send success to elk
		msg := fmt.Sprintf("Successfully deployed to %v", address)
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

	log.L.Debugf("Sending event to %v", os.Getenv("ELASTIC_API_EVENTS"))
	elkreporting.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS"), e, 3*time.Second)

	return report
}
