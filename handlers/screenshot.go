package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/labstack/echo"
)

func GetScreenshot(context echo.Context) error {
	hostname := context.Param("hostname")
	img, err := helpers.MakeScreenshot(hostname)

	if err != nil {
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
