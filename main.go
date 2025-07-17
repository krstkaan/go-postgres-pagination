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

	// VM endpoints
	vmRepo := repository.NewVMRepository(database)
	vmSvc := service.NewVMService(vmRepo)
	vmHandler := handler.NewVMHandler(vmSvc)

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// New VM endpoints
	router.GET("/api/vms", vmHandler.GetVMs)
	router.GET("/api/vms/search", vmHandler.SearchVMs)
	router.GET("/api/vms/:id", vmHandler.GetVMByID)

	router.Run("0.0.0.0:5000")
}
