package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(api *gin.RouterGroup) {
	healthHandler := handlers.NewHealthHttpHandler()
	api.GET("/health", healthHandler.GetHealth)
}
