package helpers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

	output := make(map[int]string)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	for _, piece := range bytes.Split(body, []byte("$$$")) {

		reader := bufio.NewReader(bytes.NewReader(piece))

		rawId, err := reader.ReadSlice(byte('\n'))
		if err != nil {
			return nil, err
		}

		log.Printf("%s", color.HiMagentaString("rawId: %s", string(rawId)))

		toConvert := bytes.Trim(rawId, " \n")

		log.Printf("%s", color.HiMagentaString("toConvert: %s", string(toConvert)))

		id, err := strconv.Atoi(string(toConvert))
		if err != nil {
			return nil, err
		}

		fileName, err := GenerateRandomString(NUM_BYTES)
		if err != nil {
			return nil, err
		}

		fileLocation := os.Getenv("GOPATH") + DOCKER_PATH

		outFile, err := os.OpenFile(fileLocation+fileName, os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(outFile, reader)
		if err != nil {
			return nil, err
		}

		outFile.Close()
		TrackFile(fileName, fileLocation)

		output[id] = fileName

	}

	log.Printf("%v", output)

	return output, nil
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
