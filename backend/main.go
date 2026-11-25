package main

import (
	"log"

	"exam-system-mvp/config"
	"exam-system-mvp/handlers"
	"exam-system-mvp/models"
	"exam-system-mvp/observability"
	"exam-system-mvp/router"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if err := observability.Configure(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.File); err != nil {
		log.Fatalf("Failed to configure application logger: %v", err)
	}
	gin.SetMode(cfg.Server.Mode)

	// 1. Connect to the database
	db := cfg.ConnectDatabase()

	// 2. Run migrations
	models.Migrate(db)

	// 3. Seed initial data
	models.SeedData(db)

	// 4. Auto-submit in-progress exams when their server-side deadline expires.
	go handlers.StartAttemptAutoSubmitWorker(db)

	// 5. Setup and run the HTTP server
	r := router.SetupRouterWithConfig(db, cfg)

	port := cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
