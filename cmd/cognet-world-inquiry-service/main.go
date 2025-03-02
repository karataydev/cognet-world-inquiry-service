package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"cognet-world-inquiry-service/internal/config"
	"cognet-world-inquiry-service/internal/handler"
	"cognet-world-inquiry-service/internal/service"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisAddress,
		Password: config.AppConfig.RedisPassword,
	})

	// Verify Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize services
	dataImporter := service.NewDataImporter(redisClient)
	cognateSearchService := service.NewCognateSearch(redisClient)

	// Initialize handlers
	importHandler := handler.NewImportHandler(dataImporter)

	cognateHandler := handler.NewCognateHandler(cognateSearchService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:      handler.ErrorHandler,
		BodyLimit:         0,               // No limit
		ReadBufferSize:    1024 * 1024 * 4, // 4MB buffer
		WriteBufferSize:   1024 * 1024 * 4, // 4MB buffer
		StreamRequestBody: true,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Setup routes
	setupRoutes(app, importHandler, cognateHandler)

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + config.AppConfig.ServerPort); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	log.Printf("Server started on port %s", config.AppConfig.ServerPort)

	// Wait for interrupt signal
	<-shutdownChan
	log.Println("Shutting down server...")

	// Cleanup and shutdown
	if err := app.Shutdown(); err != nil {
		log.Fatal("Server shutdown error:", err)
	}
}

func setupRoutes(app *fiber.App, importHandler *handler.ImportHandler, cognateHandler *handler.CognateHandler) {
	api := app.Group("/api/v1")

	// Import routes
	importRoutes := api.Group("/import")
	importRoutes.Post("/tsv", importHandler.ImportTSV)
	importRoutes.Post("/languages", importHandler.ImportLanguages)
	importRoutes.Get("/status", importHandler.GetStatus)
	importRoutes.Delete("/clear", importHandler.ClearDatabase)

	// Search routes
	searchRoutes := api.Group("/search")
	searchRoutes.Get("/suggestions", cognateHandler.GetSuggestions)
	searchRoutes.Get("/concept/:id", cognateHandler.GetByConceptID)
	searchRoutes.Get("/chains/concept/:id", cognateHandler.FindCognateChains)

}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
