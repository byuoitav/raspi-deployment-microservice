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

	err := helpers.Deploy(hostname, []byte{}, []byte{}, os.Stdout)
	if err != nil {
		log.L.Warnf(err.Error())
	}

	return ctx.JSON(http.StatusOK, nil)
}

// DeployByTypeAndDesignation handles the echo request to deploy to a type/designation
func DeployByTypeAndDesignation(ctx echo.Context) error {
	deviceType := ctx.Param("type")
	deviceDesignation := ctx.Param("designation")

	/*
		err := helpers.Deploy("itb-1101-cp2.byu.edu", "stage", []byte{}, []byte{})
		if err != nil {
			log.L.Warnf(err.Addf("failed to deploy to", deviceType).Error())
		}
	*/

	/*
		response, err := helpers.ScheduleDeployment(deviceClass, deploymentType)
		if err != nil {
			context.JSON(http.StatusBadRequest, err.Error())
			return nil
		}
	*/

	return ctx.JSON(http.StatusOK, nil)
}

// DeployByBuildingAndTypeAndDesignation handles the echo request to deploy to a building/type/designation
func DeployByBuildingAndTypeAndDesignation(ctx echo.Context) error {
	return nil
}
