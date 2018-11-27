package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/common/log"
)

// NumBytes .
// const NumBytes = 8

// Port .
// const Port = ":5001" // port the designation microservice works on

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

// GetClassAndDesignationID .
func GetClassAndDesignationID(class, designation string) (int64, int64, error) {
	if (len(class) == 0) || (len(designation) == 0) {
		return 0, 0, errors.New("invalid class or designation")
	}

	//get class ID
	classID, err := GetClassId(class)
	if err != nil {
		msg := fmt.Sprintf("class ID not found: %s", err.Error())
		//		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	//get designation ID
	desigID, err := GetDesignationId(designation)
	if err != nil {
		msg := fmt.Sprintf("designation ID not found: %s", err.Error())
		//		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	return classID, desigID, nil
}

// MakeEnvironmentRequest .
func MakeEnvironmentRequest(endpoint string) (*http.Response, error) {
	var client http.Client

	url := os.Getenv("DESIGNATION_MICROSERVICE_ADDRESS") + endpoint

	//	log.Printf("[helplers] making request against url %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &http.Response{}, fmt.Errorf("unable to request docker-compose or etc/environment file: %s", err.Error())
	}

	err = SetToken(req)
	if err != nil {
		return &http.Response{}, fmt.Errorf("unable to request docker-compose or etc/environment file: %s", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

// SetToken .
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

// GetServiceFromS3 .
func GetServiceFromS3(service, designation string) ([]file, bool, error) {
	files := []file{}
	serviceFileExists := false

	log.L.Infof("Getting files in s3 from %s/%s", designation, service)
	objects, err := GetS3Folder(os.Getenv("AWS_BUCKET_REGION"), os.Getenv("AWS_S3_SERVICES_BUCKET"), fmt.Sprintf("%s/device-monitoring", designation))
	if err != nil {
		return nil, serviceFileExists, fmt.Errorf("unable to download s3 service %s (designation: %s): %s", service, designation, err)
	}

	for name, bytes := range objects {
		file := file{
			Path:  fmt.Sprintf("/byu/%s/%s", service, name),
			Bytes: bytes,
		}

		if name == service {
			file.Permissions = 0100
		} else if name == fmt.Sprintf("%s.service.tmpl", service) {
			serviceFileExists = true
			file.Permissions = 0644
		} else {
			file.Permissions = 0644
		}

		log.L.Debugf("added file %v, permissions %v", file.Path, file.Permissions)
		files = append(files, file)
	}

	log.L.Infof("Successfully got %v files.", len(files))
	return files, serviceFileExists, nil
}

// GetS3Folder .
func GetS3Folder(region, bucket, prefix string) (map[string][]byte, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region: aws.String(region),
	})

	// get list of objects
	listObjectsResp, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get s3 folder: %v", err)
	}

	// build a downloader for s3
	downloader := s3manager.NewDownloaderWithClient(svc)

	wg := sync.WaitGroup{}
	objects := make(map[string][]byte)
	objectsMu := sync.Mutex{}
	errors := []error{}

	for _, key := range listObjectsResp.Contents {
		log.L.Debugf("Downloading %v from bucket %v", *key.Key, bucket)
		wg.Add(1)

		go func(key *string) {
			var bytes []byte
			buffer := aws.NewWriteAtBuffer(bytes)
			_, err := downloader.Download(buffer, &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    key,
			})
			if err != nil {
				errors = append(errors, err)
			}

			name := strings.TrimPrefix(*key, prefix)
			name = strings.TrimPrefix(name, "/")

			objectsMu.Lock()
			objects[name] = buffer.Bytes()
			objectsMu.Unlock()

			wg.Done()
		}(key.Key)
	}
	wg.Wait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors downloading folder from s3: %s", errors)
	}

	return objects, nil
}
