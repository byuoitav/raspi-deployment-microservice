package helpers

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/fatih/color"
)

//const FILE_NAME = "environment-variables"                       //name of file we use later
const PORT = ":5001"                                            //port the designation microservice works on
const ENDPOINT = "/configurations/designations/%d/%d/variables" //endpoint we use to make request against designation microservice

// retrieveEnvironmentVariables gets the environment variables for each Pi as a file to SCP over
func retrieveEnvironmentVariables(class, designation string) (string, error) {

	log.Printf("[helpers] fetching environment variables...")

	if (len(class) == 0) || (len(designation) == 0) {
		return "", errors.New("invalid class or designation")
	}

	//get class ID
	classId, err := GetClassId(class)
	if err != nil {
		msg := fmt.Sprintf("class ID not found: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	//get designation ID
	desigId, err := GetDesignationId(designation)
	if err != nil {
		msg := fmt.Sprintf("designation ID not found: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	var client http.Client
	url := os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS") + fmt.Sprintf(ENDPOINT, classId, desigId)

	log.Printf("[helplers] making request against url %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		msg := fmt.Sprintf("cannot make new request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	err = SetToken(req)
	if err != nil {
		msg := fmt.Sprintf("failed to set bearer token: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to execute request: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("non-200 response from designation microservice: %d", resp.StatusCode)
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	fileName := fmt.Sprintf("%s-%s", class, designation)

	fileLocation := os.Getenv("GOPATH") + "/src/github.com/byuoitav/raspi-deployment-microservice/public/"
	log.Printf("[helpers] filepath: %s", fileLocation)
	outFile, err := os.OpenFile(fileLocation+fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	outFile.Close()

	testFile, err := os.Open(fileLocation + fileName)
	if err != nil {
		return "", errors.New(fmt.Sprintf("can't open test file: %s", err.Error()))
	}

	contents, err := ioutil.ReadAll(testFile)
	if err != nil {
		return "", err
	}

	log.Printf("[helpers] contents of %s: %s", fileLocation+fileName, string(contents))

	testFile.Close()

	return fileName, nil
}

func SetToken(request *http.Request) error {

	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {

		log.Printf("[helplers] setting bearer token...")

		token, err := bearertoken.GetToken()
		if err != nil {
			msg := fmt.Sprintf("cannot get bearer token: %s", err.Error())
			log.Printf("%s", color.HiRedString("[helpers] %s", msg))
			return errors.New(msg)
		}

		request.Header.Set("Authorization", "Bearer "+token.Token)
	}

	return nil
}
