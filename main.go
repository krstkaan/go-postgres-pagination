package main

import (
	"go-postgres-pagination/db"
	"go-postgres-pagination/handler"
	"go-postgres-pagination/repository"
	"go-postgres-pagination/service"

	"github.com/gin-gonic/gin"
)

func main() {
	database := db.Init()

	repo := repository.NewServerRepository(database)
	svc := service.NewServerService(repo)
	h := handler.NewServerHandler(svc)

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	router.GET("/api/servers", h.GetServers)
	router.GET("/api/servers/search", h.SearchServers) // Yeni eklenen route
	router.GET("/api/servers/:id", h.GetServerByID)    // ID'ye göre server getiren route

	router.Run("0.0.0.0:5000")

}
