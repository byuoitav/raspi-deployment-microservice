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
	log.L.Infof("We are entering GetScreenshot!")
	address := context.Request().RemoteAddr
	log.L.Infof(address)
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.L.Infof("Failed to read Request body: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	log.L.Infof("%s", body)

	//	err = context.Request().ParseForm()
	if err != nil {
		log.L.Infof("Failed to Parse Form: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}
	//	log.L.Infof("%s", context.Request().PostForm)
	//	text := context.Request().PostFormValue("&text")
	sections := strings.Split(string(body), "&text=")

	textSection := strings.Split(sections[1], "&")
	text := textSection[0]
	text = text + ".byu.edu"

	userSection := strings.Split(sections[0], "&user_name=")
	userSection = strings.Split(userSection[1], "&")
	userName := userSection[0]

	channelSection := strings.Split(sections[0], "&channel_id=")
	channelSection = strings.Split(userSection[1], "&")
	channelID := channelSection[0]

	log.L.Infof(text)
	err = helpers.MakeScreenshot(text, address, userName, channelID)

	if err != nil {
		log.L.Infof("Failed to MakeScreenshot: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	log.L.Infof("We are exiting GetScreenshot")

	return context.JSON(http.StatusOK, "Screenshot confirmed")
}

func ReceiveScreenshot(context echo.Context) error {
	log.L.Infof("I have entered ReceiveScreenshot")
	ScreenshotName := context.Param("ScreenshotName")
	img, err := ioutil.ReadAll(context.Request().Body)
	defer context.Request().Body.Close()
	if err != nil {
		log.L.Errorf("Could not read in the screenshot")
		return context.JSON(http.StatusInternalServerError, err)
	}
	//0644 is the OS.FileMode
	err = ioutil.WriteFile("/tmp/"+ScreenshotName+".xwd", img, 0644)
	if err != nil {
		log.L.Errorf("Could not write out the screenshot")
		return context.JSON(http.StatusInternalServerError, err)
	}
	log.L.Infof("We are finishing receiving the screenshot")
	return context.JSON(http.StatusOK, "Hooray!")
}
