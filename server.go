package main

import (
	"log"
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/hateoas"
	"github.com/byuoitav/raspi-deployment-microservice/handlers"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	err := hateoas.Load("https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/swagger.json")
	if err != nil {
		log.Fatalln("Could not load Swagger file. Error: " + err.Error())
	}

	port := ":8008"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	router.Static("/*", "public")
	router.GET("/", echo.WrapHandler(http.HandlerFunc(hateoas.RootResponse)))
	router.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))

	secure.GET("/webhook/:designation", handlers.DeployDesignation) //	looks for all rooms with the given designation and deploys to all roles

	secure.GET("/webhook/designations/:designation/roles/:role", handlers.DeployDesignationByRole) //	targets all devices with the given role in all rooms

	secure.GET("/webhook/rooms/:room/:role", handlers.DeployRoomByDesignationAndRole) //	targets all devices in the given room with the given role

	secure.GET("/webhook_device/:hostname", handlers.WebhookDevice) //	targets a specific device

	secure.GET("/webhook_contacts/enable/:hostname", handlers.EnableContacts) //	TODO figure out how to handle physical monitoring
	secure.GET("/webhook_contacts/disable/:hostname", handlers.DisableContacts)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
