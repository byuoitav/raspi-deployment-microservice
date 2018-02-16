package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/jessemillar/jsonresp"
	"github.com/labstack/echo"
)

func DeployDesignation(context echo.Context) error {

	return context.JSON(http.StatusNotImplemented, "not implemented")
}

func DeployDesignationByRole(context echo.Context) error {

	designation := context.Param("designation")
	role := context.Param("role")

	err := helpers.Deploy(designation, role)
	if err != nil {
		return context.JSON(http.StatusBadRequest, err)
	}

	return context.JSON(http.StatusOK, fmt.Sprintf("%s %s deployment started", designation, role))
}

func DeployRoomByDesignationAndRole(context echo.Context) error {

	return context.JSON(http.StatusNotImplemented, "not implemented")
}

func WebhookDevice(context echo.Context) error {
	response, err := helpers.DeployDevice(context.Param("hostname"))
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
