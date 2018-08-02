package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/raspi-deployment-microservice/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := ":8008"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	//

	// unautheticated routes
	router.Static("/*", "public")
	router.GET("/health", health)
	router.GET("/mstatus", mstatus)

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	/* secure routes */
	// deployment
	secure.GET("/webhook_device/:hostname", handlers.DeployByHostname)
	secure.GET("/webhook/:type/:designation", handlers.DeployByTypeAndDesignation)
	secure.GET("/webhook_building/:building/:type/:designation", handlers.DeployByBuildingAndTypeAndDesignation)

	// divider sensor contacts enable/disable
	secure.GET("/webhook_contacts/enable/:hostname", handlers.EnableContacts)
	secure.GET("/webhook_contacts/disable/:hostname", handlers.DisableContacts)

	// TODO websocket and ui endpoints
	// websocket/ui

	err := router.StartServer(&http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	})
	if err != nil {
		log.L.Errorf("failed to start http server: %v", err)
	}
}

func health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}

// TODO
func mstatus(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}

/*
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
*/
