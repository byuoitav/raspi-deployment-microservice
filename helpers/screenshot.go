package helpers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/raspi-deployment-microservice/socket"
	"golang.org/x/crypto/ssh"
)

var sshConfiguration *ssh.ClientConfig

//Builds the sshConfig
func init() {
	// get ssh key
	bucket := s3.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("AWS_BUCKET_REGION")),
	})

	resp, err := bucket.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("RASPI_DEPLOYMENT_S3_BUCKET")),
		Key:    aws.String(os.Getenv("AWS_DEPLOYMENT_KEY")),
	})
	if err != nil {
		log.L.Fatalf("failed to get aws deployment key")
	}
	defer resp.Body.Close()
	log.L.Infof("Successfully got AWS deployment key.")
	// read key from response
	key, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Fatalf("unable to read private ssh key: %v", err)
	}

	// parse the pem encoded private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.L.Fatalf("unable to read parse private ssh key: %v", err)
	}

	// get pi username
	uname := os.Getenv("PI_SSH_USERNAME")
	if len(uname) == 0 {
		log.L.Fatalf("PI_SSH_USERNAME must be set.")
	}

	// build ssh config
	sshConfiguration = &ssh.ClientConfig{
		User: uname,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO should we check the host key..?
		Timeout:         5 * time.Second,
	}
}

func MakeScreenshot(hostname string, address string) ([]byte, error) {
	img := []byte{}
	//Make our ssh client, writer, and sesh

	client, err := ssh.Dial("tcp", hostname+":22", sshConfiguration)
	if err != nil {
		return img, nerr.Translate(err).Addf("Client could not be created")
	}
	output := socket.Writer(hostname)
	if err != nil {
		return img, nerr.Translate(err).Addf("Ssh sesh could not be opened")
	}

	defer client.Close()
	//Try to Open Session
	sesh, er := NewSession(client, output)
	if er != nil {
		msg := fmt.Sprintf("unable to open sesh with %v: %v", hostname, er.Error())
		fmt.Fprintf(output, msg)
		return img, er
	}
	//Try to Open (Warp) Pipe
	stdin, err := sesh.StdinPipe()
	if err != nil {
		msg := fmt.Sprintf("unable to open stdin pipe on %v: %v", hostname, err)
		fmt.Fprintf(output, msg)
		return img, err
	}

	//Try to Create a (Koopa) Shell
	err = sesh.Shell()
	if err != nil {
		msg := fmt.Sprintf("unable to start shell on %v: %v", output, err)
		fmt.Fprintf(output, msg)
		return img, err
	}

	log.L.Debugf("Started new shell on %s", hostname)
	ScreenshotName := hostname + "*" + time.Now().Format(time.RFC3339)
	fmt.Fprintf(stdin, `script -f /tmp/screenshot.log`+"\n")
	//Take the Screenshot
	fmt.Fprintf(stdin, `xwd -out %s.xwd -root -display :0.0`+"\n", ScreenshotName)
	//Puts the Screenshot onto AWS
	fmt.Fprintf(stdin, `curl -XPOST https://byuoitav-raspi-deployment-microservice.avs.byu.edu/ReceiveScreenshot/%s -T ./%s.xwd`+"\n", ScreenshotName, ScreenshotName)
	//Remove the Screenshot
	fmt.Fprintf(stdin, `rm %s.xwd`+"\n", ScreenshotName)
	fmt.Fprintln(stdin, `exit`)
	stdin.Close()
	err = sesh.Wait()
	if err != nil {
		log.L.Warnf("failed to screenshot %v: %v", hostname, err)
	}
	//Convert the Screenshot to a .png
	time.Sleep(250 * time.Millisecond)
	FullScreenshotName := fmt.Sprintf("/tmp/%s.xwd", ScreenshotName)
	cmd := exec.Command("convert", FullScreenshotName, ScreenshotName+".png")
	cmd.Run()
	//	cmd = exec.Command("rm", FullScreenshotName)
	//	cmd.Run()
	//Read in the Screenshot
	//img, err = ioutil.ReadFile("screenshot.png")

	if err != nil {
		log.L.Infof("Failed to read Screenshot file %v: %v", ScreenshotName, err)
	}

	//TODO Put the screenshot into a public s3 bucket and then send that url back thus pulling the picture down to slack
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String("byu-oit-av-screenshot-bucket"), //TODO CHANGE BUCKET
		Key:           aws.String(ScreenshotName),
		Body:          bytes.NewReader(img),
		ContentLength: aws.Int64(int64(len(img))), //Size of Image
		ContentType:   aws.String(".png"),
	})

	if err != nil {
		log.L.Infof("Everything about Amazon has failed: %v", err)
		return img, err
	}

	return img, nil
}
