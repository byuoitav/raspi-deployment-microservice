package helpers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

func GetDeviceDocker(target structs.Device) (string, error) {

	log.Printf("[helpers] requesting docker file for: %s", target.Name)

	url := fmt.Sprintf("%s/docker/devices/%d", os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"), target.ID)

	log.Printf("[helpers] making request to: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	err = SetToken(req)
	if err != nil {
		return "", err
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("non-200 response from designation microservice: %d", resp.StatusCode))
	}

	var fileName string

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &fileName)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

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

	finder := regexp.MustCompile("[0-9]+")

	for _, piece := range bytes.Split(body, []byte("$$$")) {

		rawId := finder.Find(piece)
		if rawId == nil {
			log.Printf("%s", color.HiRedString("unable to find device ID"))
			continue
		}

		reader := bufio.NewReader(bytes.NewReader(piece))

		_, err := reader.Discard(len(rawId) + 1) //	discard the id and the newline from the buffer
		if err != nil {
			log.Printf("%s", color.HiRedString("error discarding: %s", err.Error()))
			continue
		}

		log.Printf("%s", color.HiMagentaString("rawId: %s", string(rawId)))

		id, err := strconv.Atoi(string(rawId))
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
