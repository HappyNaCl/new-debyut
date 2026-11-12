package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(
	api *gin.RouterGroup,
	authClient auth.AuthServiceClient,
) {
	authHandler := handlers.NewAuthHttpHandler(authClient)
	api.POST("/auth/register", authHandler.RequestRegisterOtp)
	api.POST("/auth/register/verify", authHandler.VerifyRegisterOtp)
	api.POST("/auth/register/complete", authHandler.CompleteRegistration)
	api.POST("/auth/login", authHandler.RequestLoginOtp)
	api.POST("/auth/login/verify", authHandler.VerifyLoginOtp)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	protected := api.Group("/auth", middleware.JwtAuth())
	protected.POST("/logout", authHandler.Logout)
}
