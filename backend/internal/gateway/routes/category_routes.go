package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(
	api *gin.RouterGroup,
	eventClient event.EventServiceClient,
) {
	categoryHandler := handlers.NewCategoryHttpHandler(eventClient)
	api.GET("/categories", categoryHandler.List)

	admin := api.Group("/admin/categories", middleware.JwtAuth(), middleware.RequireRole(auth.Role_ADMIN.String()))
	admin.POST("", categoryHandler.Create)
	admin.PUT("/:id", categoryHandler.Update)
	admin.DELETE("/:id", categoryHandler.Delete)
}
