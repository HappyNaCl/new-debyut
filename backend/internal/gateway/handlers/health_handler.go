package handlers

import (
	"net/http"
	"time"

	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/health"
	"github.com/gin-gonic/gin"
)

type HealthHttpHandler struct {
	startTime time.Time
}

func NewHealthHttpHandler() *HealthHttpHandler {
	return &HealthHttpHandler{
		startTime: time.Now(),
	}
}

// @Summary      Service health and uptime
// @Tags         Health
// @Produce      json
// @Success      200  {object}  dto.HealthResponse
// @Router       /health [get]
func (h HealthHttpHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: http.StatusOK,
		Uptime: time.Since(h.startTime).Round(time.Second).String(),
	})
}
