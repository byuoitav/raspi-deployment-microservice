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
	log.L.Infof("WE MADE IT!")
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
	sections = strings.Split(sections[1], "&")
	text := sections[0]
	text = text + ".byu.edu"
	log.L.Infof(text)
	img, err := helpers.MakeScreenshot(text, address)

	if err != nil {
		log.L.Infof("Failed to MakeScreenshot: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	return context.Blob(http.StatusOK, "image/png", img)
}

func ReceiveScreenshot(context echo.Context) error {
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

	return context.JSON(http.StatusOK, "Hooray!")
}
