package routes

import (
	"github.com/HappyNaCl/new-debyut/internal/gateway/handlers"
	"github.com/HappyNaCl/new-debyut/internal/gateway/middleware"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(
	api *gin.RouterGroup,
	organizationClient users.OrganizationServiceClient,
) {
	organizationHandler := handlers.NewOrganizationHttpHandler(organizationClient)

	organizations := api.Group("/organizations", middleware.JwtAuth())
	organizations.GET("/me", organizationHandler.ListMine)
	organizations.GET("/:id", organizationHandler.Get)
	organizations.PUT("/:id", organizationHandler.Update)
	organizations.PUT("/:id/avatar", organizationHandler.UpdateAvatar)
	organizations.GET("/:id/members", organizationHandler.ListMembers)
	organizations.DELETE("/:id/members/:userId", organizationHandler.RemoveMember)
	organizations.PUT("/:id/members/:userId/status", organizationHandler.SetMemberStatus)
	organizations.POST("/:id/invitations", organizationHandler.CreateInvitation)
	organizations.GET("/:id/invitations", organizationHandler.ListInvitations)
	organizations.DELETE("/:id/invitations/:invitationId", organizationHandler.RevokeInvitation)

	invitations := api.Group("/invitations", middleware.JwtAuth())
	invitations.GET("/me", organizationHandler.ListMyInvitations)
	invitations.POST("/:id/accept", organizationHandler.AcceptInvitation)
	invitations.POST("/:id/decline", organizationHandler.DeclineInvitation)

	admin := api.Group("/admin/organizations", middleware.JwtAuth(), middleware.RequireRole(auth.Role_ADMIN.String()))
	admin.GET("", organizationHandler.List)
}
