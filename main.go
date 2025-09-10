package main

import (
	"claude2api/config"
	"claude2api/router"
	"claude2api/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Initialize all services (config is already loaded in init())
	service.InitServices(config.ConfigInstance)

	// Setup all routes
	router.SetupRoutes(r)

	// Run the server on 0.0.0.0:8080
	r.Run(config.ConfigInstance.Address)
}
