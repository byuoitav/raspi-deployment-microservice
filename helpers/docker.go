package helpers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

//builds a map of device IDs to docker-compose files
func GetRoomDocker(room structs.Room, role string) (map[int]string, error) {

	roleId, err := GetConfigDbRoleId(role)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/configurations/rooms/%d/roles/%d", os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"), room.ID, roleId)

	log.Printf("[helpers] making request against %s for room: %s", color.HiCyanString(url), color.HiGreenString(room.Name))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		msg := fmt.Sprintf("cannot make new request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return nil, errors.New(msg)
	}

	err = SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set bearer token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return nil, errors.New(msg)
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("non-200 response from designation microservice: %d", resp.StatusCode)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return nil, errors.New(msg)
	}

	reader := bufio.NewReader(resp.Body)
	scanner := bufio.NewScanner(reader)

	log.Printf("[helpers] processing designation data...")
	scanner.Split(split)

	for scanner.Scan() {

		log.Printf("result of split: %s", scanner.Text())
	}

	return nil, nil
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {

	if atEOF {
		return 0, nil, nil
	}

	delim := []byte("$$$")

	yamls := bytes.SplitN(data, delim, 2)

	if (len(yamls[0]) + len(delim)) > len(data) {

		var toReturn []byte
		copy(toReturn, data)

		return 0, toReturn, nil
	}

	return (len(yamls[0]) + len(delim)), yamls[0], nil
}

func GetConfigDbRoleId(role string) (int, error) {

	roles, err := dbo.GetDeviceRoleDefinitions()
	if err != nil {
		return 0, err
	}

	for _, possibleRole := range roles {

		if strings.Compare(possibleRole.Name, role) == 0 {

			return possibleRole.ID, nil
		}
	}

	return 0, errors.New("invalid role")
}
