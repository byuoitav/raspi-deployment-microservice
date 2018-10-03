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
	"github.com/nlopes/slack"
	"golang.org/x/crypto/ssh"
)

/*
type slackResponse struct {
	Token      string            `json:"token"`
	Channel    string            `json:"channel"`
	Text       string            `json:"text"`
	Attachment map[string]string `json:"attachment"`
}
*/

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

func MakeScreenshot(hostname string, address string, userName string, outputChannelID string) error {

	img := []byte{}

	//Make our ssh client, writer, and sesh

	client, err := ssh.Dial("tcp", hostname+":22", sshConfiguration)
	if err != nil {
		return nerr.Translate(err).Addf("Client could not be created")
	}

	output := socket.Writer(hostname)
	if err != nil {
		return nerr.Translate(err).Addf("Ssh sesh could not be opened")
	}

	defer client.Close()

	//Try to Open Session
	sesh, er := NewSession(client, output)
	if er != nil {
		msg := fmt.Sprintf("unable to open sesh with %v: %v", hostname, er.Error())
		fmt.Fprintf(output, msg)
		return er
	}

	//Try to Open (Warp) Pipe
	stdin, err := sesh.StdinPipe()
	if err != nil {
		msg := fmt.Sprintf("unable to open stdin pipe on %v: %v", hostname, err)
		fmt.Fprintf(output, msg)
		return err
	}

	//Try to Create a (Koopa) Shell
	err = sesh.Shell()
	if err != nil {
		msg := fmt.Sprintf("unable to start shell on %v: %v", output, err)
		fmt.Fprintf(output, msg)
		return err
	}

	log.L.Debugf("Started new shell on %s", hostname)
	//Our Screenshot name is a combination of the hostname and a timestamp to make it unique
	ScreenshotName := hostname + "*" + time.Now().Format(time.RFC3339)
	//Log our adventures on the pi
	fmt.Fprintf(stdin, `script -f /tmp/screenshot.log`+"\n")
	//Take the Screenshot
	fmt.Fprintf(stdin, `xwd -out %s.xwd -root -display :0.0`+"\n", ScreenshotName)
	//Puts the Screenshot onto AWS
	fmt.Fprintf(stdin, `curl -XPOST https://byuoitav-raspi-deployment-microservice.avs.byu.edu/ReceiveScreenshot/%s -T ./%s.xwd`+"\n", ScreenshotName, ScreenshotName)
	//Remove the Screenshot
	fmt.Fprintf(stdin, `rm %s.xwd`+"\n", ScreenshotName)
	fmt.Fprintln(stdin, `exit`)
	//Close the Pipe and the Session
	stdin.Close()
	err = sesh.Wait()
	if err != nil {
		log.L.Warnf("failed to screenshot %v: %v", hostname, err)
	}
	//Convert the Screenshot to a .png with a moment to make sure that the screenshot is posted first
	time.Sleep(250 * time.Millisecond)
	FullScreenshotName := fmt.Sprintf("/tmp/%s.xwd", ScreenshotName)
	cmd := exec.Command("convert", FullScreenshotName, ScreenshotName+".png")
	err = cmd.Run()
	if err != nil {
		log.L.Errorf("Failed to execute convert command: %v", err.Error())
	}

	//TODO Do we want to remove the screenshot from the docker?
	//	cmd = exec.Command("rm", FullScreenshotName)
	//	cmd.Run()

	//Read in the Screenshot
	img, err = ioutil.ReadFile(ScreenshotName + ".png")

	if err != nil {
		log.L.Infof("Failed to read Screenshot file %v: %v", ScreenshotName, err.Error())
	}

	//Puts the Picture into the s3 Bucket
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(os.Getenv("SLACK_AHOY_BUCKET")),
		Key:           aws.String(ScreenshotName), //Image Name
		Body:          bytes.NewReader(img),       //The Image
		ContentLength: aws.Int64(int64(len(img))), //Size of Image
		ContentType:   aws.String(".png"),
	})

	if err != nil {
		log.L.Infof("Everything about Amazon has failed: %v", err)
		return err
	}
	//New Slack thing with token
	myToken := os.Getenv("SLACK_AHOY_TOKEN")
	api := slack.New(myToken)

	//Initialize Paramters and Create them
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:     "Here is " + userName + "'s screenshot of " + hostname,
		ImageURL: "http://s3-us-west-2.amazonaws.com/" + os.Getenv("SLACK_AHOY_BUCKET") + "/" + ScreenshotName,
	}

	//Make the Parameters Official
	params.Attachments = []slack.Attachment{attachment}
	//Post the Message to Slack
	channelID, timestamp, err := api.PostMessage(outputChannelID, "Ahoy!", params)

	if err != nil {
		log.L.Errorf("We failed to send to Slack: %s", err.Error())
	}
	//Log if we succeeded and where we succeeded
	log.L.Infof("Message successfully sent to channel %s at %s", channelID, timestamp)

	log.L.Infof("We made it to the end boys. It is done.")
	return nil
}
