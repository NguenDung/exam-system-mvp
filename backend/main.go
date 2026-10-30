package main

import (
	"log"

	"exam-system-mvp/config"
	"exam-system-mvp/handlers"
	"exam-system-mvp/models"
	"exam-system-mvp/router"
)

func main() {
	config.ConnectDatabase()
	models.Migrate(config.DB)
	models.SeedData(config.DB)

	go handlers.StartAttemptAutoSubmitWorker()

	r := router.SetupRouter()

	port := config.GetEnv("SERVER_PORT", "8081")
	log.Printf("Server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
