package handlers

import (
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/labstack/echo"
)

type circlePayload struct {
	Name string `json:"reponame"`
}

func Webhook(context echo.Context) error {
	repo := circlePayload{}
	err := context.Bind(&repo)
	if err != nil {
		return context.String(http.StatusBadRequest, err.Error())
	}

	response, err := helpers.Deploy(repo.Name)
	if err != nil {
		return context.String(http.StatusBadRequest, err.Error())
	}

	return context.JSON(http.StatusOK, response)
}
