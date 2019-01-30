package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/byuoitav/common/log"
)

type Attachment struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
}

type Message struct {
	Token       string       `json:"token"`
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

// MakeScreenshot takes a screenshot on device and posts it to a slack channel
func MakeScreenshot(hostname string, address string, userName string, outputChannelID string) error {
	img := []byte{}

	resp, err := http.Get("http://" + hostname + ":10000/device/screenshot")

	if err != nil {
		log.L.Errorf("[Screenshot] We failed to get the screenshot: %s", err.Error())
	}

	log.L.Debugf("[Screenshot] Response: %v", resp)
	ScreenshotName := hostname + "*" + time.Now().Format(time.RFC3339)

	defer resp.Body.Close()
	img, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Errorf("[Screenshot] We couldn't read in the response body: %s", err)
	}

	//Puts the Picture into the s3 Bucket
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(os.Getenv("SLACK_AHOY_BUCKET")),
		Key:           aws.String(ScreenshotName), //Image Name
		Body:          bytes.NewReader(img),       //The Image
		ContentLength: aws.Int64(int64(len(img))), //Size of Image
		ContentType:   aws.String(".jpg"),
	})

	if err != nil {
		log.L.Errorf("[Screenshot] Everything about Amazon has failed: %v", err)
		return err
	}
	//New Slack thing with token
	myToken := os.Getenv("SLACK_AHOY_TOKEN")
	myWebHook := os.Getenv("SLACK_AHOY_WEBHOOK")

	attachment := Attachment{
		Title:    "Here is " + userName + "'s screenshot of " + hostname,
		ImageURL: "http://s3-us-west-2.amazonaws.com/" + os.Getenv("SLACK_AHOY_BUCKET") + "/" + ScreenshotName,
	}

	var attachments []Attachment
	attachments = append(attachments, attachment)

	message := Message{
		Token:       myToken,
		Channel:     outputChannelID,
		Text:        "Ahoy!",
		Attachments: attachments,
	}

	//Marshal it
	j, err := json.Marshal(message)
	if err != nil {
		log.L.Errorf("[Screenshot] failed to marshal message: %v", message)
		return err
	}

	//Make the request
	req, err := http.NewRequest("POST", myWebHook, bytes.NewBuffer(j))
	req.Header.Set("Content-type", "application/json")

	slackClient := &http.Client{}
	resp, err = slackClient.Do(req)

	if err != nil {
		log.L.Errorf("[Screenshot] We failed to send to Slack: %s", err.Error())
		return err
	}

	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Errorf("[Screenshot] Couldn't read the Slack Response: %s", err.Error())
		return err
	}
	log.L.Warnf("[Screenshot] Slack Response: %s", p)

	if resp.StatusCode/100 != 2 {
		log.L.Error("[Screenshot] Non-200 Response")
		return errors.New(string(p))
	}

	var i interface{}
	err = json.Unmarshal(p, i)
	if err != nil {
		log.L.Errorf("[Screenshot] Couldn't unmarshal: %s", err.Error())
		return err
	}

	log.L.Warnf("[Screenshot] Unmarshalled body: %+v", i)

	log.L.Warnf("We made it to the end boys. It is done.")
	return nil
}
