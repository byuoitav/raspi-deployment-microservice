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

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"golang.org/x/crypto/ssh"
)

// DeployReport is returned after attempting a deployment
type DeployReport struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"msg"`
	Success   bool   `json:"success"`
}

var (
	signer     ssh.Signer
	piUsername = os.Getenv("PI_SSH_USERNAME")
)

func init() {
	// read private key file
	key, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")
	if err != nil {
		log.L.Fatalf("unable to read private ssh key: %v", err)
	}

	// parse the pem encoded private key
	signer, err = ssh.ParsePrivateKey(key)
	if err != nil {
		log.L.Fatalf("unable to read parse private ssh key: %v", err)
	}

	// validate $PI_SSH_USERNAME is set
	if len(piUsername) == 0 {
		log.L.Fatalf("PI_SSH_USERNAME must be set.")
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

	reports, err = deployHelper([]structs.Device{device}, device.Type.ID, room.Designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to deploy to device %v", device.ID)
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

	reports, err = deployHelper(allDevices, deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to deploy to devices by type %v and designation %v", deviceType, designation)
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

	reports, err = deployHelper(allDevices, deviceType, designation)
	if err != nil {
		return reports, nerr.Translate(err).Addf("failed to deploy to devices in building %v by type %v and designation %v", building, deviceType, designation)
	}

	return reports, nil
}

// DeployHelper takes a slice of devices and gets all the data it needs to deploy
func deployHelper(devices []structs.Device, deviceType, designation string) ([]DeployReport, *nerr.E) {
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
		go func() {
			report := deploy(devices[i].Address, []byte(envVars), []byte(dockerCompose), os.Stdout)

			reportsMu.Lock()
			reports = append(reports, report)
			reportsMu.Unlock()

			wg.Done()
		}()
	}

	wg.Wait()
	return reports, nil
}

// Deploy deploys to a single pi
func deploy(address string, envVars, dockerCompose []byte, output io.Writer) DeployReport {
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

	// build ssh config
	sshConfig := &ssh.ClientConfig{
		User: piUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO should we check the host key..?
		Timeout:         5 * time.Second,
	}

	// ssh into address, on port 22
	client, err := ssh.Dial("tcp", address+":22", sshConfig)
	if err != nil {
		report.Message = fmt.Sprintf("failed to ssh into %v: %v", address, err)
		return report
	}
	defer client.Close()

	log.L.Infof("Successfully opened connection to %s", address)

	// open a new session with the client
	session, err := client.NewSession()
	if err != nil {
		report.Message = fmt.Sprintf("unable to create session with %v: %v", address, err)
		return report
	}

	// get stdout/err pipes
	stdout, err := session.StdoutPipe()
	if err != nil {
		report.Message = fmt.Sprintf("failed to get stdout pipe with %v: %v", address, err)
		return report
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		report.Message = fmt.Sprintf("failed to get stderr pipe with %v: %v", address, err)
		return report
	}

	// read from output pipes, and write the output to output
	go readWrite("stdout", stdout, output, 512*1)
	go readWrite("stderr", stderr, output, 512*1)

	// execute command
	err = session.Run("docker ps")
	if err != nil {
		report.Message = fmt.Sprintf("failed to run command on %v: %v", address, err)
		return report
	}

	return report
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

/*
type device struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

var TIMER_DURATION = 3 * time.Minute

func reportToELK(hostname string, msg string, success bool) elkReport {

	var key string

	if success {
		key = "Successful"
	} else {
		key = "Failed"
	}

	report := elkReport{Hostname: hostname, Timestamp: time.Now().Format(time.RFC3339), Message: key + ": " + msg}

	splitName := strings.Split(hostname, "-")

	var e events.Event

	if len(splitName) != 3 {
		e = events.Event{
			Hostname:         hostname,
			Timestamp:        time.Now().Format(time.RFC3339),
			LocalEnvironment: false,
			Building:         "",
			Room:             "",
			Event: events.EventInfo{
				Type:           events.DEPLOYMENT,
				Requestor:      "",
				EventCause:     events.AUTOGENERATED,
				Device:         hostname,
				EventInfoKey:   key,
				EventInfoValue: msg,
			},
		}
	} else {
		e = events.Event{
			Hostname:         hostname,
			Timestamp:        time.Now().Format(time.RFC3339),
			LocalEnvironment: false,
			Building:         splitName[0],
			Room:             splitName[0] + "-" + splitName[1],
			Event: events.EventInfo{
				Type:           events.DEPLOYMENT,
				Requestor:      "",
				EventCause:     events.AUTOGENERATED,
				Device:         splitName[2][:strings.Index(splitName[2], ".")],
				EventInfoKey:   key,
				EventInfoValue: msg,
			},
		}
	}

	log.L.Debugf("Sending event to %v", os.Getenv("ELASTIC_API_EVENTS"))
	elkreporting.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS"), e, 3*time.Second)

	return report
}
*/
