package helpers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/byuoitav/common/db"

	"github.com/byuoitav/common/structs"
)

func DeploySpanelByHostname(hostname string) (string, error) {
	allCaps := strings.ToUpper(hostname)
	log.Printf("[%s] Starting deployment", allCaps)

	room, err := db.GetDB().GetRoom(allCaps)
	if err != nil {
		msg := fmt.Sprintf("[%s] failed to get room: %s", allCaps, err)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	deviceName := strings.Split(allCaps, "-")[2]

	if !strings.Contains(deviceName, "SP") {
		msg := fmt.Sprintf("[%s] device name must match 'SPX'", allCaps)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	var device structs.Device

	for _, d := range room.Devices {
		if d.Name == deviceName {
			device = d
			break
		}
	}

	if len(device.Name) == 0 {
		msg := fmt.Sprintf("[%s] couldn't find device %s in room %s", allCaps, deviceName, room.Name)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	envFile, err := retrieveEnvironmentVariables("SchedulingPanel", "s-dev")
	if err != nil {
		msg := fmt.Sprintf("[%s] failed to get environment variables: %s", allCaps, err)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	dockerCompose, err := RetrieveDockerCompose("SchedulingPanel", "s-dev")
	if err != nil {
		msg := fmt.Sprintf("[%s] failed to get docker-compose file: %s", allCaps, err)
		log.Printf(msg)
		return "", errors.New(msg)
	}

	respChannel := make(chan elkReport, 1)
	go SendCommand(device.Address, envFile, dockerCompose, respChannel)

	<-respChannel

	msg := fmt.Sprintf("[%s] deployment started", allCaps)
	return msg, nil
}
