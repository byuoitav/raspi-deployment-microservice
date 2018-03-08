package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/fatih/color"
	"github.com/labstack/echo"
)

func DeployDesignation(context echo.Context) error {

	return context.JSON(http.StatusNotImplemented, "not implemented")
}

func DeployDesignationByRole(context echo.Context) error {

	designation := context.Param("designation")
	role := context.Param("role")

	err := helpers.DeployDesignation(designation, role)
	if err != nil {
		return context.JSON(http.StatusBadRequest, err)
	}

	return context.JSON(http.StatusOK, fmt.Sprintf("%s %s deployment started", designation, role))
}

func DeployRoomByDesignationAndRole(context echo.Context) error {

	room := context.Param("room")
	role := context.Param("role")

	err := helpers.DeployRoom(room, role)
	if err != nil {
		msg := fmt.Sprintf("deployment to %ss in %s failed: %s", role, room, err.Error())
		log.Printf("%s", color.HiRedString("[handlers] %s", msg))
		return context.JSON(http.StatusBadRequest, msg)
	}

	return context.JSON(http.StatusOK, fmt.Sprintf("started deployment to %ss in %s", role, room))
}

func WebhookDevice(context echo.Context) error {

	hostname := context.Param("hostname")
	device, err := helpers.GetDevice(context.Param("hostname"))
	if err != nil {
		msg := fmt.Sprintf("device %s not found: %s", hostname, err.Error())
		log.Printf("%s", color.HiRedString("[handlers] %s", msg))
		return context.JSON(http.StatusBadRequest, msg)
	}

	err = helpers.DeployDevice(device)
	if err != nil {
		msg := fmt.Sprintf("deployment to %s failed: %s", hostname, err.Error())
		log.Printf("%s", color.HiRedString("[handlers] %s", msg))
		return context.JSON(http.StatusBadRequest, msg)
	}

	return context.JSON(http.StatusOK, fmt.Sprintf("successfully deployed to: %s", hostname))
}
