package main

import (
	"log"

	"exam-system-mvp/config"
	"exam-system-mvp/models"
	"exam-system-mvp/router"
)

func main() {
	// 1. Connect to the database
	config.ConnectDatabase()

	// 2. Run migrations
	models.Migrate(config.DB)

	// 3. Seed initial data
	models.SeedData(config.DB)

	// 4. Setup and run the HTTP server
	r := router.SetupRouter()

	port := config.GetEnv("SERVER_PORT", "8081")
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
