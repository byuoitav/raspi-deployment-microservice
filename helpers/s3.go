package helpers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/common/log"
)

//const FILE_NAME = "environment-variables"                       //name of file we use later
const NUM_BYTES = 8
const PORT = ":5001"                                            // port the designation microservice works on
const ENDPOINT = "/configurations/designations/%d/%d/variables" // endpoint we use to make request against designation microservice

var (
	filePath string
)

func init() {
	ex, err := os.Executable()
	if err != nil {
		log.L.Fatalf("Failed to get location of executable: %v", err)
	}

	filePath = filepath.Dir(ex)
}

// retrieveEnvironmentVariables gets the environment variables for each Pi as a file to SCP over
//func retrieveEnvironmentVariables(class, designation string) (string, error) {
func retrieveEnvironmentVariables(class, designation string) ([]byte, error) {
	var resp []byte

	//	log.Printf("[helpers] fetching environment variables...")

	classId, desigId, err := GetClassAndDesignationID(class, designation)
	if err != nil {
		return resp, errors.New(fmt.Sprintf("invalid class or designation: %s", err.Error()))
	}

	response, err := MakeEnvironmentRequest(fmt.Sprintf("/configurations/designations/%d/%d/variables", classId, desigId))
	if err != nil {
		return resp, err
	}

	if response.StatusCode != http.StatusOK {
		msg, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return resp, errors.New(fmt.Sprintf("non-200 response from pi-designation-microservice: %d, unable to read response: %s", response.StatusCode, err.Error()))
		}
		return resp, errors.New(fmt.Sprintf("non-200 response from pi-designation-microservice: %d, message: %s", response.StatusCode, string(msg)))
	}

	b, err := ioutil.ReadAll(response.Body)
	return b, err
}

//func RetrieveDockerCompose(class, designation string) (string, error) {
func RetrieveDockerCompose(class, designation string) ([]byte, error) {
	var bytes []byte

	//	log.Printf("[helpers] retrieving docker-compose file for devices of class: %s, designation: %s", class, designation)

	//get class and designation IDs
	classID, desigId, err := GetClassAndDesignationID(class, designation)
	if err != nil {
		return bytes, errors.New(fmt.Sprintf("invalid class or designation: %s", err.Error()))
	}

	resp, err := MakeEnvironmentRequest(fmt.Sprintf("/configurations/designations/%d/%d/docker-compose", classID, desigId))
	if err != nil {
		return bytes, err
	}

	if resp.StatusCode != http.StatusOK {
		return bytes, errors.New(fmt.Sprintf("non-200 response from pi-designation-microservice: %d", resp.StatusCode))
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

func GetClassAndDesignationID(class, designation string) (int64, int64, error) {

	if (len(class) == 0) || (len(designation) == 0) {
		return 0, 0, errors.New("invalid class or designation")
	}

	//get class ID
	classId, err := GetClassId(class)
	if err != nil {
		msg := fmt.Sprintf("class ID not found: %s", err.Error())
		//		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	//get designation ID
	desigId, err := GetDesignationId(designation)
	if err != nil {
		msg := fmt.Sprintf("designation ID not found: %s", err.Error())
		//		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	return classId, desigId, nil
}

func MakeEnvironmentRequest(endpoint string) (*http.Response, error) {

	var client http.Client

	url := os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS") + endpoint

	//	log.Printf("[helplers] making request against url %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &http.Response{}, errors.New(fmt.Sprintf("unable to request docker-compose or etc/environment file: %s", err.Error()))
	}

	err = SetToken(req)
	if err != nil {
		return &http.Response{}, errors.New(fmt.Sprintf("unable to request docker-compose or etc/environment file: %s", err.Error()))
	}

	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func SetToken(request *http.Request) error {

	//	log.Printf("[helpers] setting bearer token...")

	token, err := bearertoken.GetToken()
	if err != nil {
		msg := fmt.Sprintf("cannot get bearer token: %s", err.Error())
		//		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return errors.New(msg)
	}

	request.Header.Set("Authorization", "Bearer "+token.Token)

	return nil
}
