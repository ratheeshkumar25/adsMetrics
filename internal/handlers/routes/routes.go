package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/ratheeshkumar25/adsmetrictracker/docs"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/handlers"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/middleware"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	handler *handlers.Handler
}

func NewRouter(handler *handlers.Handler) *Router {
	return &Router{
		handler: handler,
	}
}

func (r *Router) SetupRoutes(log *logger.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(log))
	router.Use(r.corsMiddleware())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimiter())

	// Health and system endpoints
	r.setupSystemRoutes(router)

	// API v1 routes - ONLY REQUIRED ENDPOINTS
	r.setupAPIRoutes(router)

	return router
}

func (r *Router) setupSystemRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", r.handler.Health)

	// Ready check (for Kubernetes)
	router.GET("/ready", r.handler.Ready)

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation with proper configuration
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("doc.json"),
		ginSwagger.DeepLinking(true),
		ginSwagger.DocExpansion("none"),
	))

	// API info endpoint
	router.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":     "ads-metric-tracker",
			"version":     "1.0.0",
			"description": "High-performance ads metric tracking API",
			"swagger":     "/swagger/index.html",
		})
	})
}

func (r *Router) setupAPIRoutes(router *gin.Engine) {
	// Core API endpoints as per requirements
	router.GET("/ads", r.handler.GetAds)                 // R: GET /ads
	router.POST("/ads/click", r.handler.PostClick)       // R: POST /ads/click
	router.GET("/ads/analytics", r.handler.GetAnalytics) // R: GET /ads/analytics
}

func (r *Router) corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure properly for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
