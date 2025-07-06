package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/services"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/metrics"
)

type Handler struct {
	adsService *services.AdsService
	log        *logger.Logger
}

func NewHandler(adsService *services.AdsService, log *logger.Logger) *Handler {
	return &Handler{
		adsService: adsService,
		log:        log,
	}
}

// GetAds godoc
//	@Summary		Get all ads
//	@Description	Returns a list of ads with basic metadata.
//	@Tags			ads
//	@Produce		json
//	@Success		200	{object}	GetAdsResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/ads [get]
func (h *Handler) GetAds(c *gin.Context) {
	start := time.Now()

	ads, err := h.adsService.GetAdsAllAds()
	if err != nil {
		h.log.Logger.Errorf("Failed to get ads: %v", err)
		metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch ads",
			"message": err.Error(),
		})
		return
	}

	// Transform ads to response format (basic metadata only)
	response := make([]AdResponse, len(ads))
	for i, ad := range ads {
		response[i] = AdResponse{
			ID:        ad.ID,
			ImageURL:  ad.ImageURL,
			TargetURL: ad.TargetURL,
			CreatedAt: ad.CreatedAt,
		}
	}

	metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"ads":   response,
		"count": len(response),
	})
}

// PostClick godoc
//	@Summary		Record ad click event
//	@Description	Accepts a click payload and processes it asynchronously.
//	@Tags			clicks
//	@Accept			json
//	@Produce		json
//	@Param			click	body		ClickRequest	true	"Click event data"
//	@Success		202		{object}	ClickResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/ads/click [post]
func (h *Handler) PostClick(c *gin.Context) {
	start := time.Now()

	var request ClickRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Logger.Errorf("Invalid click request: %v", err)
		metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	// Validate required fields
	if request.AdID == "" {
		metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "400", time.Since(start).Seconds())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ad_id is required",
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()
	if request.IP == "" {
		request.IP = clientIP
	}

	// Create click model
	click := model.Clicks{
		ID:            uuid.New().String(),
		AdID:          request.AdID,
		IP:            request.IP,
		VideoPlayTime: request.VideoPlayTime,
		Timestamp:     time.Now(),
	}

	// Use timestamp from request if provided
	if !request.Timestamp.IsZero() {
		click.Timestamp = request.Timestamp
	}

	// Publish to Kafka for asynchronous processing (non-blocking)
	go func() {
		if err := h.adsService.PublishClick(click); err != nil {
			h.log.Logger.Errorf("Failed to publish click to Kafka: %v", err)
			// Fallback: process directly if Kafka fails
			if err := h.adsService.ProcessClick(click); err != nil {
				h.log.Logger.Errorf("Failed to process click directly: %v", err)
			}
		}
	}()

	// Return immediate response to client
	metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "202", time.Since(start).Seconds())
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Click recorded",
		"click_id":   click.ID,
		"ad_id":      click.AdID,
		"timestamp":  click.Timestamp,
		"processing": "asynchronous",
	})
}

// GetAnalytics godoc
//	@Summary		Get ad analytics
//	@Description	Returns real-time analytics for a specific ad or all ads.
//	@Tags			Analytics
//	@Produce		json
//	@Param			ad_id		query		string	false	"Filter by Ad ID"
//	@Param			timeframe	query		string	false	"Time window (1m, 5m, 15m, 1h, 24h)"
//	@Success		200			{object}	AnalyticsOverview
//	@Failure		500			{object}	ErrorResponse
//	@Router			/ads/analytics [get]
func (h *Handler) GetAnalytics(c *gin.Context) {
	start := time.Now()

	// Get query parameters
	adID := c.Query("ad_id")
	timeFrame := c.DefaultQuery("timeframe", "1h")

	// If specific ad requested
	if adID != "" {
		analytics, err := h.adsService.GetAnalytics(adID)
		if err != nil {
			h.log.Logger.Errorf("Failed to get analytics for ad %s: %v", adID, err)
			metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "500", time.Since(start).Seconds())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch analytics",
				"message": err.Error(),
			})
			return
		}

		metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "200", time.Since(start).Seconds())
		c.JSON(http.StatusOK, analytics)
		return
	}

	// Get analytics for all ads
	ads, err := h.adsService.GetAdsAllAds()
	if err != nil {
		h.log.Logger.Errorf("Failed to get ads for analytics: %v", err)
		metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "500", time.Since(start).Seconds())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch analytics",
			"message": err.Error(),
		})
		return
	}

	// Get analytics for each ad
	analyticsData := make([]services.AnalyticsResponse, 0, len(ads))
	for _, ad := range ads {
		analytics, err := h.adsService.GetAnalytics(ad.ID)
		if err != nil {
			h.log.Logger.Warnf("Failed to get analytics for ad %s: %v", ad.ID, err)
			continue
		}
		analyticsData = append(analyticsData, *analytics)
	}

	response := AnalyticsOverview{
		TotalAds:    len(ads),
		TimeFrame:   timeFrame,
		Analytics:   analyticsData,
		GeneratedAt: time.Now(),
	}

	metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), "200", time.Since(start).Seconds())
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "ads-metric-tracker",
	})
}

func (h *Handler) Ready(c *gin.Context) {
	// You can add actual readiness checks here (database connectivity, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now(),
		"service":   "ads-metric-tracker",
	})
}

// Metrics endpoint for Prometheus
func (h *Handler) Metrics(c *gin.Context) {
	// This will be handled by the prometheus handler in the router
	c.String(http.StatusOK, "Metrics endpoint")
}

// Request/Response models
type ClickRequest struct {
	AdID          string    `json:"ad_id" binding:"required"`
	IP            string    `json:"ip,omitempty"`
	VideoPlayTime int       `json:"video_play_time,omitempty"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
}

type AdResponse struct {
	ID        string    `json:"id"`
	ImageURL  string    `json:"image_url"`
	TargetURL string    `json:"target_url"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsOverview struct {
	TotalAds    int                          `json:"total_ads"`
	TimeFrame   string                       `json:"timeframe"`
	Analytics   []services.AnalyticsResponse `json:"analytics"`
	GeneratedAt time.Time                    `json:"generated_at"`
}

type GetAdsResponse struct {
	Ads   []AdResponse `json:"ads"`
	Count int          `json:"count"`
}

type ClickResponse struct {
	Message    string    `json:"message"`
	ClickID    string    `json:"click_id"`
	AdID       string    `json:"ad_id"`
	Timestamp  time.Time `json:"timestamp"`
	Processing string    `json:"processing"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Middleware for request logging and metrics
func (h *Handler) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())

		// Log request
		h.log.Logger.Infof("%s %s - %s (%v)", method, path, statusCode, duration)

		// Record metrics
		metrics.RecordHTTPRequest(method, path, statusCode, duration.Seconds())
	}
}
