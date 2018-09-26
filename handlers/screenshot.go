package handlers

import (
	"io/ioutil"
	"net/http"

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

	err = context.Request().ParseForm()
	if err != nil {
		log.L.Infof("Failed to Parse Form: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}
	log.L.Infof("%s", context.Request().PostForm)
	text := context.Request().PostFormValue("&text")
	log.L.Infof(text)
	img, err := helpers.MakeScreenshot(text, address)

	if err != nil {
		log.L.Infof("Failed to MakeScreenshot: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, err)
	}

	return context.Blob(http.StatusOK, "image/png", img)
}

/*
func SendScreenshotToSlack(context echo.Context) error {
	hostname := context.Param("hostname")
	channelID := context.Param("channelID")

	err := helpers.MakeScreenshotGoToSlack(hostname, channelID)

	if err != nil {
		return context.JSON(http.StatusInternalServerError, err)
	}

	return context.JSON(http.StatusOK, nil)

}
*/
func ReceiveScreenshot(context echo.Context) error {
	ScreenshotName := context.Param("ScreenshotName")
	img, err := ioutil.ReadAll(context.Request().Body)
	defer context.Request().Body.Close()
	if err != nil {
		//TODO I think this is wrong

		return err
	}
	//TODO -> Store img Somewhere
	//0644 is the OS.FileMode
	ioutil.WriteFile(ScreenshotName+".xwd", img, 0644)

	//TODO -> Make this return thing make sense
	return context.JSON(http.StatusOK, "Hooray!")
}
