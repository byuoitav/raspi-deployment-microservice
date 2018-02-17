package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/pi-designation-microservice/accessors"
	"github.com/fatih/color"
)

func GetRoleId(roleName string) (int64, error) {

	//	log.Printf("[helpers] getting class ID corresponding to class: %s", color.HiCyanString(roleName))

	var client http.Client
	url := fmt.Sprintf("%s/classes/definitions/all", os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS"))

	//	log.Printf("[helpers] making request against url %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		msg := fmt.Sprintf("cannot make new request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	err = SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set bearer token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to execute request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("non-200 response from designation microservice: %d", resp.StatusCode)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("unable to read response body: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	var roles []accessors.Class
	err = json.Unmarshal(body, &roles)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal class structs from JSON: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	for _, possibleRole := range roles {

		//		log.Printf("[helpers] considering role: %s", color.HiCyanString(possibleRole.Name))

		if possibleRole.Name == roleName { //found class ID
			return possibleRole.ID, nil
		}
	}

	return 0, errors.New("class not found") //if we make it this far without finding it, it wasn't there
}

func GetDesignationId(desigName string) (int64, error) {

	log.Printf("[helpers] getting designation ID corresponding to class: %s", desigName)

	var client http.Client
	url := os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS") + "/designations/definitions/all"

	log.Printf("[helplers] making request against url %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		msg := fmt.Sprintf("cannot make new request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	err = SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set bearer token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to execute request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("non-200 response from designation microservice: %d", resp.StatusCode)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("unable to read response body: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	var designations []accessors.Designation
	err = json.Unmarshal(body, &designations)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal class structs from JSON: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, errors.New(msg)
	}

	for _, designation := range designations {

		if designation.Name == desigName { //found class ID
			return designation.ID, nil
		}
	}

	return 0, errors.New("designation not found") //if we make it this far without finding it, it wasn't there
}
