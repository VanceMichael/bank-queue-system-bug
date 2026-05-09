package main

import (
	"bank-queue-system/config"
	"bank-queue-system/handlers"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitRedis()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		queue := api.Group("/queue")
		{
			queue.GET("/business-types", handlers.GetBusinessTypes)
			queue.POST("/generate", handlers.GenerateQueue)
			queue.GET("/waiting", handlers.GetWaitingQueues)
			queue.GET("/:id", handlers.GetQueueByID)
			queue.GET("/:id/position", handlers.GetQueuePosition)
			queue.GET("/wait-time/estimate", handlers.EstimateWaitTime)
		}

		window := api.Group("/window")
		{
			window.POST("", handlers.CreateWindow)
			window.GET("", handlers.GetAllWindows)
			window.GET("/:id", handlers.GetWindowByID)
			window.PUT("/:id/status", handlers.UpdateWindowStatus)
			window.PUT("/:id/business-types", handlers.UpdateWindowBusinessTypes)
			window.POST("/:id/call-next", handlers.CallNextQueue)
			window.POST("/:id/start-processing", handlers.StartProcessing)
			window.POST("/:id/complete", handlers.CompleteQueue)
			window.POST("/:id/missed", handlers.MissedQueue)
			window.POST("/recall/:id", handlers.RecallMissedQueue)
			window.POST("/vip-insert", handlers.VIPInsertQueue)
		}

		api.GET("/statistics", handlers.GetStatistics)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
