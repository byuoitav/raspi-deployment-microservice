package main

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	si "github.com/byuoitav/device-monitoring-microservice/statusinfrastructure"
	"github.com/byuoitav/raspi-deployment-microservice/handlers"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := ":8008"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	router.Static("/*", "public")
	router.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))
	router.GET("/mstatus", GetStatus)

	secure.GET("/webhook/:class/:designation", handlers.WebhookDeployment)
	secure.GET("/webhook/:branch/disable", handlers.DisableDeploymentsByBranch)
	secure.GET("/webhook/:branch/enable", handlers.EnableDeploymentsByBranch)

	secure.GET("/webhook_building/:building/:class/:designation", handlers.WebhookDeploymentByBuilding)

	secure.GET("/webhook_device/:hostname", handlers.WebhookDevice)
	secure.GET("/webhook_contacts/enable/:hostname", handlers.EnableContacts)
	secure.GET("/webhook_contacts/disable/:hostname", handlers.DisableContacts)

	secure.GET("/webhook/scheduling/:designation", handlers.WebhookSchedulingDeployment)
	secure.GET("/webhook_device/scheduling/:hostname", handlers.WebhookSchedulingDevice)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}

func GetStatus(context echo.Context) error {
	var s si.Status
	var err error

	s.Version, err = si.GetVersion("version.txt")
	if err != nil {
		return context.JSON(http.StatusOK, "Failed to open version.txt")
	}

	// Test a database retrieval to assess the status.
	vals, err := db.GetDB().GetAllBuildings()
	if len(vals) < 1 || err != nil {
		s.Status = si.StatusDead
		s.StatusInfo = fmt.Sprintf("Unable to access database. Error: %s", err)
	} else {
		s.Status = si.StatusOK
		s.StatusInfo = ""
	}
	log.L.Info("Getting Mstatus")

	return context.JSON(http.StatusOK, s)
}
