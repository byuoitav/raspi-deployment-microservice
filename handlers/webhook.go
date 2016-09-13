package handlers

import (
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/jessemillar/jsonresp"
	"github.com/labstack/echo"
)

func Webhook(context echo.Context) error {
	response, err := helpers.Deploy()
	if err != nil {
		return jsonresp.New(context, http.StatusBadRequest, err.Error())
	}

	return jsonresp.New(context, http.StatusOK, response)
}
