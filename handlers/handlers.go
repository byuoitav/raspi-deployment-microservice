package handlers

import (
	"net/http"
	"os"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/raspi-deployment-microservice/helpers"
	"github.com/labstack/echo"
)

// DeployByHostname handles the echo request to deploy to a single device
func DeployByHostname(ctx echo.Context) error {
	hostname := ctx.Param("hostname")

	reports, err := helpers.DeployByHostname(hostname)
	if err != nil {
		log.L.Warnf(err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.String())
	}

	return ctx.JSON(http.StatusOK, reports)
}

/*
// DeployByTypeAndDesignation handles the echo request to deploy to a type/designation
func DeployByTypeAndDesignation(ctx echo.Context) error {
	deviceType := ctx.Param("type")
	deviceDesignation := ctx.Param("designation")

	reports, err := helpers.DeployByTypeAndDesignation(deviceType, deviceDesignation)
	if err != nil {
		log.L.Warnf(err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.String())
	}

	return ctx.JSON(http.StatusOK, reports)
}

// DeployByBuildingAndTypeAndDesignation handles the echo request to deploy to a building/type/designation
func DeployByBuildingAndTypeAndDesignation(ctx echo.Context) error {
	building := ctx.Param("building")
	deviceType := ctx.Param("type")
	deviceDesignation := ctx.Param("designation")

	reports, err := helpers.DeployByBuildingAndTypeAndDesignation(building, deviceType, deviceDesignation)
	if err != nil {
		log.L.Warnf(err.Error())
		return ctx.JSON(http.StatusInternalServerError, err.String())
	}

	return ctx.JSON(http.StatusOK, reports)
}
*/

// EnableContacts handles the echo request to enable the contacts service on a specific hostname
func EnableContacts(context echo.Context) error {
	report, err := helpers.UpdateContactState(context.Param("hostname"), true, os.Stdout)
	if err != nil {
		log.L.Warnf(err.Error())
		return context.JSON(http.StatusInternalServerError, err.String())
	}

	return context.JSON(http.StatusOK, report)
}

// DisableContacts handles the echo request to disable the contacts service on a specific hostname
func DisableContacts(context echo.Context) error {
	report, err := helpers.UpdateContactState(context.Param("hostname"), false, os.Stderr)
	if err != nil {
		log.L.Warnf(err.Error())
		return context.JSON(http.StatusInternalServerError, err.String())
	}

	return context.JSON(http.StatusOK, report)
}
