package helpers

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// retrieveEnvironmentVariables gets the environment variables for eacah Pi as a file to SCP over
func retrieveEnvironmentVariables() (string, error) {
	fileName := "environment-variables"
	fileLocation := os.Getenv("GOPATH") + "/src/github.com/byuoitav/raspi-deployment-microservice/public/"

	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	response, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("elasticbeanstalk-us-west-2-194925301021"),
		Key:    aws.String("environment-variables"),
	})

	if err != nil {
		return fileName, err
	}

	defer response.Body.Close()

	outFile, err := os.OpenFile(fileLocation+fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fileName, err
	}

	defer outFile.Close()

	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
