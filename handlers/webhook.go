package handlers

import (
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/jessemillar/jsonresp"
	"github.com/labstack/echo"
)

func WebhookDeployment(context echo.Context) error {
	branch := context.Param("branch")
	response, err := helpers.ScheduleDeployment(branch)
	if err != nil {
		jsonresp.New(context.Response(), http.StatusBadRequest, err.Error())
		return nil
	}

	jsonresp.New(context.Response(), http.StatusOK, response)
	return nil
}

func WebhookDevice(context echo.Context) error {
	response, err := helpers.DeploySingle(context.Param("hostname"))
	if err != nil {
		jsonresp.New(context.Response(), http.StatusBadRequest, err.Error())
		return nil
	}

	jsonresp.New(context.Response(), http.StatusOK, response)
	return nil
}

func EnableContacts(context echo.Context) error {

	err := helpers.UpdateContactState(context.Param("hostname"), true)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, map[string]string{"Response": "Failed to set state"})
	}

	return context.JSON(http.StatusOK, map[string]string{"Response": "Success"})
}

func DisableContacts(context echo.Context) error {

	err := helpers.UpdateContactState(context.Param("hostname"), false)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, map[string]string{"Response": "Failed to set state"})
	}

	return context.JSON(http.StatusOK, map[string]string{"Response": "Success"})
}
