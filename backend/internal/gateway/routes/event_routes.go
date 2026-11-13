package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

func RegisterEventRoutes(
	api *gin.RouterGroup,
	eventClient event.EventServiceClient,
	organizationClient users.OrganizationServiceClient,
) {
	eventHandler := handlers.NewEventHttpHandler(eventClient, organizationClient)

	api.GET("/events", eventHandler.ListLatest)
	api.GET("/events/:id", eventHandler.GetPublic)

	events := api.Group("/organizations/:id/events", middleware.JwtAuth())
	events.GET("", eventHandler.List)
	events.POST("", eventHandler.Create)
	events.POST("/images", eventHandler.UploadImage)
	events.GET("/:eventId", eventHandler.Get)
	events.PUT("/:eventId", eventHandler.Update)
}
