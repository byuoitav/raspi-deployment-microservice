package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
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
