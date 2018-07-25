package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/labstack/echo"
)

func WebhookDeployment(context echo.Context) error {
	deviceClass := context.Param("class")
	deploymentType := context.Param("designation")

	response, err := helpers.ScheduleDeployment(deviceClass, deploymentType)
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return nil
	}

	context.JSON(http.StatusOK, response)
	return nil
}

func DisableDeploymentsByBranch(context echo.Context) error {
	branch := context.Param("branch")
	helpers.HoldDeployment(branch, true)
	return context.String(http.StatusOK, fmt.Sprintf("Disabled %s deployments", branch))
}

func EnableDeploymentsByBranch(context echo.Context) error {
	branch := context.Param("branch")
	helpers.HoldDeployment(branch, false)
	return context.String(http.StatusOK, fmt.Sprintf("Enabled %s deployments", branch))
}

func WebhookDevice(context echo.Context) error {
	response, err := helpers.DeployDevice(context.Param("hostname"))
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return nil
	}

	context.JSON(http.StatusOK, response)
	return nil
}

func WebhookSchedulingDeployment(context echo.Context) error {
	return nil
}

func WebhookSchedulingDevice(context echo.Context) error {
	hostname := context.Param("hostname")
	response, err := helpers.DeploySpanelByHostname(hostname)
	if err != nil {
		log.Printf("error deploying spanel by hostname: %s", err)
		return err
	}

	context.JSON(http.StatusOK, response)
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
