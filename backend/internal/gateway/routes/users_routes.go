package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(
	api *gin.RouterGroup,
	userClient users.UserServiceClient,
) {
	userHandler := handlers.NewUserHttpHandler(userClient)

	protectedMe := api.Group("/me", middleware.JwtAuth())
	protectedMe.GET("", userHandler.GetMe)
	protectedMe.PUT("", userHandler.UpdateMe)
	protectedMe.PUT("/avatar", userHandler.UpdateAvatar)
}
