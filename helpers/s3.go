package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/pi-designation-microservice/accessors"
	"github.com/fatih/color"
)

const FILE_NAME = "environment-variables" //name of file we use later
const PORT = ":5001"                      //port the designation microservice works on
const ENDPOINT = "/variables/%s"          //endpoint we use to make request against designation microservice

// retrieveEnvironmentVariables gets the environment variables for each Pi as a file to SCP over
func retrieveEnvironmentVariables(designation string) (string, error) {

	log.Printf("[helpers] fetching environment variables...")

	var client http.Client
	url := os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS") + fmt.Sprintf(ENDPOINT, designation)

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response body: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	var variables []accessors.Variable
	err = json.Unmarshal(body, &variables)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal environment variables: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	var buffer bytes.Buffer
	for _, variable := range variables {

		buffer.WriteString(variable.Key + "=" + variable.Value + "\n")
	}

	fileLocation := os.Getenv("GOPATH") + "/src/github.com/byuoitav/raspi-deployment-microservice/public/"
	outFile, err := os.OpenFile(fileLocation+FILE_NAME, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}

	defer outFile.Close()

	content := strings.NewReader(buffer.String())
	_, err = io.Copy(outFile, content)
	if err != nil {
		return "", err
	}

	return FILE_NAME, nil
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
