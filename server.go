package main

import (
	"io/ioutil"
	"net/http"

	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/raspi-deployment-microservice/deploy"
	"github.com/byuoitav/raspi-deployment-microservice/handlers"
)

// TODO rename to deployment service
func main() {
	port := ":10000"
	router := common.NewRouter()

	bytes, er := deploy.GenerateDockerCompose("development")
	if er != nil {
		log.L.Fatalf("error: %s", er.Error())
	}

	log.L.Infof("Compose:\n%s\n", bytes)
	err := ioutil.WriteFile("gen-docker-compose.yml", bytes, 0644)
	if err != nil {
		log.L.Fatalf("error: %s", err)
	}

	// endpoints for raspi to login to the pi and manage the update of it
	// managed := router.Group("/force", auth.AuthorizeRequest("SOMETHING", "SOMETHING_ELSE", auth.LookupResourceFromAddress))
	router.GET("/deploy/id/:id", handlers.DeployToID)
	router.GET("/deploy/group/:group/", handlers.DeployToGroup)

	// endpoints where the pi's will check in and update if they are allowed to
	// automated := router.Group("/auto", auth.AuthorizeRequest("SOMETHING", "SOMETHING_ELSE", auth.LookupResourceFromAddress))

	err = router.StartServer(&http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	})
	if err != nil {
		log.L.Errorf("failed to start http server: %v", err)
	}
}

/*
func main() {
	port := ":8008"
	// unautheticated routes
	router.GET("/health", health)

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	/* secure routes
	// deployment
	secure.GET("/webhook_device/:hostname", handlers.DeployByHostname)
	secure.GET("/webhook/:type/:designation", handlers.DeployByTypeAndDesignation)
	secure.GET("/webhook_building/:building/:type/:designation", handlers.DeployByBuildingAndTypeAndDesignation)

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
*/
