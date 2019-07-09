package helpers

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	mapset "github.com/deckarep/golang-set"
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

// retrieveEnvironmentVariables gets the environment variables for each Pi as a file to SCP over
func retrieveEnvironmentVariables(deviceType, designation string) (map[string]string, error) {
	myMap := make(map[string]string)
	deviceInfo, err := db.GetDB().GetDeviceDeploymentInfo(deviceType)
	if err != nil {
		return myMap, err
	}
	desigDevice := deviceInfo.Designations[designation]
	for _, service := range desigDevice.Services {
		resp, err := MakeEnvironmentRequest(service, designation)
		if err != nil {
			return myMap, err
		}

		for k, v := range resp {
			myMap[k] = v
		}
	}
	for k, v := range desigDevice.EnvironmentVariables {
		myMap[k] = v
	}
	return myMap, nil
}

func addMap(a, b map[string]interface{}) error {
	var s string
	set := mapset.NewSet(s)
	for k1 := range a {
		for k2, v2 := range b {
			if k1 == k2 {
				a[k1] = v2
				set.Add(k1)
			}
		}
	}
	for k, v := range b {
		if !set.Contains(k) {
			a[k] = v
		}
	}
	return nil
}

func substituteEnvironment(byter *bytes.Buffer, arrayV []interface{}, service string, tabCount int, envMap map[string]string) error {
	byter.WriteString("\n")
	for _, listItem := range arrayV {
		for i := 0; i < tabCount; i++ {
			byter.WriteString("   ")
		}
		strVersion := listItem.(string)
		values := strings.Split(strVersion, "=$")
		str := fmt.Sprintf("  - %s=%s\n", values[0], envMap[values[1]])
		//		str := fmt.Sprintf("\t- %s=%s\n", values[0], values[1])
		byter.WriteString(str)
	}
	return nil
}

func writeServiceMap(byter *bytes.Buffer, myMap map[string]interface{}, tabCount int, service string, envMap map[string]string) error {
	for k, v := range myMap {
		for i := 0; i < tabCount; i++ {
			byter.WriteString("   ")
		}
		s := fmt.Sprintf("%s:", k)
		byter.WriteString(s)
		_, ok := v.(string)
		if ok {
			str := fmt.Sprintf(" %s\n", v)
			byter.WriteString(str)
		}
		_, ok = v.([]interface{})
		if ok {
			//If we have environment variables, do the appropriate substitution
			arrayV := v.([]interface{})
			if k == "environment" {
				substituteEnvironment(byter, arrayV, service, tabCount, envMap)
			} else {
				byter.WriteString("\n")

				for _, listItem := range arrayV {
					for i := 0; i < tabCount; i++ {
						byter.WriteString("   ")
					}
					strVersion := listItem.(string)
					str := fmt.Sprintf("  - %s\n", strVersion)
					byter.WriteString(str)
				}
			}

		}
		_, ok = v.(map[string]interface{})
		if ok {
			newMap := v.(map[string]interface{})
			byter.WriteString("\n")
			writeServiceMap(byter, newMap, (tabCount + 1), service, envMap)
		}
	}
	return nil
}

func writeMap(byter *bytes.Buffer, myMap map[string]interface{}, tabCount int, designation string, deviceType string) error {
	for k, v := range myMap {
		for i := 0; i < tabCount; i++ {
			byter.WriteString("   ")
		}
		s := fmt.Sprintf("%s:", k)
		byter.WriteString(s)
		_, ok := v.(string)
		if ok {
			str := fmt.Sprintf(" %s\n", v)
			byter.WriteString(str)
		}
		_, ok = v.([]interface{})
		if ok {
			arrayV := v.([]interface{})
			byter.WriteString("\n")

			for _, listItem := range arrayV {
				for i := 0; i < tabCount; i++ {
					byter.WriteString("   ")
				}
				strVersion := listItem.(string)
				str := fmt.Sprintf("  - %s\n", strVersion)
				byter.WriteString(str)
			}
		}
		_, ok = v.(map[string]interface{})
		if ok {
			newMap := v.(map[string]interface{})
			byter.WriteString("\n")

			resp, err := MakeEnvironmentRequest(k, designation)
			if err != nil {
				return err
			}
			deviceInfo, err := db.GetDB().GetDeviceDeploymentInfo(deviceType)
			desigDevice := deviceInfo.Designations[designation]
			for k, v := range desigDevice.EnvironmentVariables {
				resp[k] = v
			}
			writeServiceMap(byter, newMap, (tabCount + 1), k, resp)
		}
	}
	return nil
}

//RetrieveDockerCompose .
func RetrieveDockerCompose(deviceType, designation string) ([]byte, error) {
	var b []byte
	var byter bytes.Buffer
	deviceInfo, err := db.GetDB().GetDeviceDeploymentInfo(deviceType)
	if err != nil {
		log.L.Warnf("Couldn't get the %s %s out of the database", designation, deviceType)
		return b, err
	}
	desigDevice := deviceInfo.Designations[designation]
	m := make(map[string]interface{})
	for _, service := range desigDevice.Services {
		resp, err := MakeDockerRequest(service, designation)
		if err != nil {
			log.L.Warnf("Couldn't get the docker info for %s:%s", service, designation)
			return b, err
		}
		tempM := make(map[string]interface{})
		tempM[service] = resp
		addMap(m, tempM)
	}
	addMap(m, desigDevice.DockerInfo)
	byter.WriteString("version: '3'\n")
	byter.WriteString("services:\n")
	writeMap(&byter, m, 1, designation, deviceType)

	return byter.Bytes(), nil
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
func MakeEnvironmentRequest(serviceID, designation string) (map[string]string, error) {
	resp, err := db.GetDB().GetDeploymentInfo(serviceID)
	toReturn := resp.CampusConfig[designation].EnvironmentVariables
	return toReturn, err
}

// MakeDockerRequest .
func MakeDockerRequest(serviceID, designation string) (map[string]interface{}, error) {
	resp, err := db.GetDB().GetDeploymentInfo(serviceID)
	toReturn := resp.CampusConfig[designation].DockerInfo
	return toReturn, err
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

// GetServiceFromCouch .
func GetServiceFromCouch(service, designation, deviceType, deviceID string) ([]file, bool, error) {
	files := []file{}
	serviceFileExists := false
	log.L.Infof("Getting files in Couch from %s/%s", designation, service)

	objects, err := GetCouchServiceFiles(service, designation, deviceType, deviceID)
	if err != nil {
		return nil, serviceFileExists, fmt.Errorf("unable to download service %s (designation: %s) from couch: %s", service, designation, err)
	}

	for name, bytes := range objects {
		file := file{
			Path:  fmt.Sprintf("/byu/%s/%s", service, name),
			Bytes: bytes,
		}
		log.L.Debugf("Service Name: %s\n", name)
		if name == service {
			file.Permissions = 0100
		} else if name == fmt.Sprintf("%s.service", service) {
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

func serviceTemplateEnvSwap(value string, envMap map[string]string, deviceID string) string {
	if value == "$SYSTEM_ID" {
		return deviceID
	}
	if strings.Contains(value, "$") {
		cleanValue := strings.Split(value, "$")
		return envMap[cleanValue[1]]
	}
	return value

}

// I've basically given up on giving good names to these functions

func writeServiceTemplate(byter *bytes.Buffer, serviceConfig structs.ServiceConfig, deviceType, designation, deviceID string) error {
	envMap, err := retrieveEnvironmentVariables(deviceType, designation)
	if err != nil {
		return err
	}
	for k, v := range serviceConfig.Data {
		byter.WriteString(fmt.Sprintf("[%s]\n", k))
		for key, value := range v {
			if isEnvironment := strings.Split(key, "="); len(isEnvironment) == 2 {
				byter.WriteString(fmt.Sprintf("%s=%s\n", key, serviceTemplateEnvSwap(value, envMap, deviceID)))
			} else {
				byter.WriteString(fmt.Sprintf("%s=%s\n", key, value))
			}
		}
		byter.WriteString("\n")
	}
	return nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// GetCouchServiceFiles .
func GetCouchServiceFiles(service, designation, deviceType, deviceID string) (map[string][]byte, error) {
	objects := make(map[string][]byte)

	//Handle Service Template
	toFill, err := db.GetDB().GetServiceInfo(service)
	if err != nil {
		log.L.Warnf("Couldn't get the service data from Couch: %v", err)
		return objects, err
	}
	serviceConfig := toFill.Designations[designation]
	var byter bytes.Buffer
	writeServiceTemplate(&byter, serviceConfig, deviceType, designation, deviceID)
	objects[fmt.Sprintf("%v.service", service)] = byter.Bytes()

	//Handle Binary
	binary, err := db.GetDB().GetServiceAttachment(service, designation)
	if err != nil {
		log.L.Warnf("Couldn't get the binary from couch for %v-%v: %v", service, designation, err)
		return objects, err
	}
	objects[fmt.Sprintf("%v", service)] = binary

	//Handle Zipped Files
	zippy, err := db.GetDB().GetServiceZip(service, designation)
	if err != nil {
		log.L.Warnf("Couldn't get the zip file from couch: %v", err)
		return objects, err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zippy), int64(len(zippy)))
	if err != nil {
		log.L.Warnf("Couldn't open zip reader: %v", err)
		return objects, err
	}

	// Read all the files from zip archive
	for _, zipFile := range zipReader.File {
		log.L.Infof("Reading file: %v", zipFile.Name)
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			log.L.Warnf("Couldn't read zipped file: %v, %v", zipFile.Name, err)
			continue
		}
		objects[zipFile.Name] = unzippedFileBytes

	}

	return objects, nil
}

// GetServiceFromS3 .
func GetServiceFromS3(service, designation string) ([]file, bool, error) {
	files := []file{}
	serviceFileExists := false

	log.L.Infof("Getting files in s3 from %s/%s", designation, service)
	objects, err := GetS3Folder(os.Getenv("AWS_BUCKET_REGION"), os.Getenv("AWS_S3_SERVICES_BUCKET"), fmt.Sprintf("%s/device-monitoring", designation))
	if err != nil {
		return nil, serviceFileExists, fmt.Errorf("unable to download  service %s (designation: %s): %s", service, designation, err)
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
