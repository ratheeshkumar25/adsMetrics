package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ratheeshkumar25/adsmetrictracker/config"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/di"
)

// @title			Ads Metric Tracker API
// @version		1.0
// @description	A high-performance, scalable ads metric tracking system built with Go.
// @description
// @description	This API provides endpoints for:
// @description	- Fetching ads with basic metadata
// @description	- Recording ad click events with resilient, non-blocking processing
// @description	- Retrieving real-time analytics and performance metrics
// @description
// @description				## Key Features
// @description				- **High Throughput**: Handles concurrent requests under high traffic
// @description				- **Data Resilience**: No data loss with fallback mechanisms
// @description				- **Real-time Analytics**: Near real-time performance metrics
// @description				- **Scalable Architecture**: Built with microservices patterns
// @description				- **Production Ready**: Docker containerized with monitoring
//
// @termsOfService				http://swagger.io/terms/
//
// @contact.name				API Support
// @contact.url				https://github.com/ratheeshkumar25/adsmetrictracker
// @contact.email				support@adsmetrictracker.com
//
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
//
// @host						localhost:8080
// @BasePath					/
//
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	// Initialize the server using the new DI format
	router, err := di.Init()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Get configuration for server setup
	cfg := config.NewConfig()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HttpPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful server shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
		log.Println("Server gracefully stopped")
	}()

	log.Printf("Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
