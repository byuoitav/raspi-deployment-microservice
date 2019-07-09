package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/raspi-deployment-microservice/handlers"
	"github.com/byuoitav/raspi-deployment-microservice/socket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := ":8008"
	router := common.NewRouter()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// unautheticated routes
	router.Static("/*", "public")
	router.GET("/health", health)

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	/* secure routes */
	// deployment
	secure.GET("/webhook_device/:hostname", handlers.DeployByHostname)
	//Unsupported Now
	/*	secure.GET("/webhook/:type/:designation", handlers.DeployByTypeAndDesignation)
		secure.GET("/webhook_building/:building/:type/:designation", handlers.DeployByBuildingAndTypeAndDesignation)
	*/

	// divider sensor contacts enable/disable
	secure.GET("/webhook_contacts/enable/:hostname", handlers.EnableContacts)
	secure.GET("/webhook_contacts/disable/:hostname", handlers.DisableContacts)

	// websocket/ui
	secure.GET("/ws", socket.EchoServeWS)

	// TODO new pi endpoint (for showing provision number thing)
	secure.GET("/newpi", handlers.NewPI)

	//Screenshots
	router.POST("/screenshot", handlers.GetScreenshot)
	//secure.GET("/screenshot/:hostname/slack/:channelID", handlers.SendScreenshotToSlack)
	router.POST("/ReceiveScreenshot/:ScreenshotName", handlers.ReceiveScreenshot)

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
