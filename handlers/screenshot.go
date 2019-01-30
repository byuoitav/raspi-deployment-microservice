package handlers

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/labstack/echo"
)

type Message struct {
	ChannelID string `json:"channel_id"`
	Text      string `json:"text"`
}

func GetScreenshot(context echo.Context) error {
	log.L.Warnf("[Screenshot] We are entering GetScreenshot!")
	address := context.Request().RemoteAddr
	log.L.Infof(address)
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.L.Warnf("[Screenshot] Failed to read Request body: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	log.L.Warnf("%s", body)

	if err != nil {
		log.L.Infof("[Screenshot] Failed to Parse Form: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	//Parse Input Body for Parameters
	sections := strings.Split(string(body), "&text=")

	//Parse for Text (which is the hostname)
	textSection := strings.Split(sections[1], "&")
	text := textSection[0]
	text = text + ".byu.edu"

	//Parse for the User Name
	userSection := strings.Split(sections[0], "&user_name=")
	userSection = strings.Split(userSection[1], "&")
	userName := userSection[0]

	//Parse for the Channel ID
	channelSection := strings.Split(sections[0], "&channel_id=")
	channelSection = strings.Split(channelSection[1], "&")
	channelID := channelSection[0]

	//Make the Screenshot
	go func() {
		err = helpers.MakeScreenshot(text, address, userName, channelID)
		if err != nil {
			log.L.Warnf("[Screenshot] Failed to MakeScreenshot: %s", err.Error())
		}
	}()
	log.L.Warnf("[Screenshot] We are exiting GetScreenshot")

	return context.JSON(http.StatusOK, "Screenshot confirmed")
}

func ReceiveScreenshot(context echo.Context) error {
	log.L.Infof("I have entered ReceiveScreenshot")
	ScreenshotName := context.Param("ScreenshotName")
	img, err := ioutil.ReadAll(context.Request().Body)
	defer context.Request().Body.Close()
	if err != nil {
		log.L.Errorf("[Screenshot] Could not read in the screenshot")
		return context.JSON(http.StatusInternalServerError, err)
	}
	//0644 is the OS.FileMode
	err = ioutil.WriteFile("/tmp/"+ScreenshotName+".xwd", img, 0644)
	if err != nil {
		log.L.Errorf("[Screenshot] Could not write out the screenshot")
		return context.JSON(http.StatusInternalServerError, err)
	}
	log.L.Infof("We are finishing receiving the screenshot")
	return context.JSON(http.StatusOK, "Hooray!")
}
