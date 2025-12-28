package url

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type createURLRequest struct {
	OriginalURL string `json:"original_url"`
	UserID      int64  `json:"user_id"` 
}

type createURLResponse struct {
	ShortURL string `json:"short_url"`
}

// POST /api/urls
func (h *Handler) CreateShortURL(c *gin.Context) {
	var req createURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	shortCode, err := h.service.CreateShortURL(req.UserID, req.OriginalURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, createURLResponse{
		ShortURL: "http://localhost:8080/" + shortCode,
	})
}

// GET /:code
func (h *Handler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

// GET /api/urls?user_id=1
func (h *Handler) ListURLs(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	urls, err := h.service.ListURLs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, urls)
}

// GET /api/urls/stats?user_id=1
func (h *Handler) UserStats(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	stats, err := h.service.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
func (h *Handler) List(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	urls, err := h.service.ListURLs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, urls)
}