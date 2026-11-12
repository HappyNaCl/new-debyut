package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

func RegisterOrganizerApplicationRoutes(
	api *gin.RouterGroup,
	applicationClient users.OrganizerApplicationServiceClient,
) {
	applicationHandler := handlers.NewOrganizerApplicationHttpHandler(applicationClient)

	applications := api.Group("/organizer-applications", middleware.JwtAuth())
	applicants := applications.Group("", middleware.RequireRole(auth.Role_USER.String(), auth.Role_ORGANIZER.String()))
	applicants.POST("/email", applicationHandler.RequestEmailOtp)
	applicants.POST("/email/verify", applicationHandler.VerifyEmailOtp)
	applicants.POST("", applicationHandler.Submit)
	applications.GET("/me", applicationHandler.ListMine)

	admin := api.Group("/admin/organizer-applications", middleware.JwtAuth(), middleware.RequireRole(auth.Role_ADMIN.String()))
	admin.GET("", applicationHandler.List)
	admin.POST("/:id/approve", applicationHandler.Approve)
	admin.POST("/:id/decline", applicationHandler.Decline)
}
