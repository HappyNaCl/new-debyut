package handlers

import (
	"net/http"

	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/organizers"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

type OrganizationHttpHandler struct {
	organizationClient users.OrganizationServiceClient
}

func NewOrganizationHttpHandler(organizationClient users.OrganizationServiceClient) *OrganizationHttpHandler {
	return &OrganizationHttpHandler{organizationClient: organizationClient}
}

// @Summary      List organizations the current user belongs to
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=[]dto.OrganizationResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /organizations/me [get]
func (h OrganizationHttpHandler) ListMine(c *gin.Context) {
	res, err := h.organizationClient.ListMyOrganizations(c, &users.ListMyOrganizationsRequest{UserId: c.GetString("userId")})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewOrganizationListResponse(res)})
}

// @Summary      Update an organization (owners only)
// @Description  Changes the organization's name and description. The email is verified at application time and cannot be changed here.
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                         true  "Organization ID"
// @Param        request  body      dto.UpdateOrganizationRequest  true  "Name and description"
// @Success      200  {object}  object{data=dto.OrganizationResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id} [put]
func (h OrganizationHttpHandler) Update(c *gin.Context) {
	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.organizationClient.UpdateOrganization(c, &users.UpdateOrganizationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		Name:           req.Name,
		Description:    req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewOrganizationResponse(res)})
}

// @Summary      Upload an organization's avatar (owners only)
// @Description  Replaces the avatar with a JPEG, PNG or WebP image up to 2 MB.
// @Tags         Organizations
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Organization ID"
// @Param        avatar  formData  file    true  "Avatar image"
// @Success      200  {object}  object{data=dto.OrganizationResponse}
// @Failure      400  {object}  object{error=object}  "Missing, too large, or unsupported image"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/avatar [put]
func (h OrganizationHttpHandler) UpdateAvatar(c *gin.Context) {
	content, ok := readAvatar(c)
	if !ok {
		return
	}

	res, err := h.organizationClient.UpdateOrganizationAvatar(c, &users.UpdateOrganizationAvatarRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		Content:        content,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewOrganizationResponse(res)})
}

// @Summary      Get an organization
// @Description  Only members can see an organization; for anyone else it does not exist.
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  object{data=dto.OrganizationResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id} [get]
func (h OrganizationHttpHandler) Get(c *gin.Context) {
	res, err := h.organizationClient.GetOrganization(c, &users.OrganizationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewOrganizationResponse(res)})
}

// @Summary      List an organization's members
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  object{data=[]dto.MemberResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/members [get]
func (h OrganizationHttpHandler) ListMembers(c *gin.Context) {
	res, err := h.organizationClient.ListMembers(c, &users.OrganizationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewMemberListResponse(res)})
}

// @Summary      Invite someone to an organization (owners only)
// @Description  Emails an invitation to join as OWNER or STAFF. The address does not need a debyut account yet. Invitations expire after 7 days; one pending invitation per email per organization.
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                       true  "Organization ID"
// @Param        request  body      dto.CreateInvitationRequest  true  "Invitee email and role"
// @Success      201  {object}  object{data=dto.InvitationResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Failure      409  {object}  object{error=string}  "This email already has a pending invitation"
// @Router       /organizations/{id}/invitations [post]
func (h OrganizationHttpHandler) CreateInvitation(c *gin.Context) {
	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.organizationClient.CreateInvitation(c, &users.CreateInvitationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		Email:          req.Email,
		Role:           users.MemberRole(users.MemberRole_value[req.Role]),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.NewInvitationResponse(res)})
}

// @Summary      List an organization's pending invitations (owners only)
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  object{data=[]dto.InvitationResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/invitations [get]
func (h OrganizationHttpHandler) ListInvitations(c *gin.Context) {
	res, err := h.organizationClient.ListInvitations(c, &users.OrganizationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewInvitationListResponse(res)})
}

// @Summary      Revoke a pending invitation (owners only)
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id            path      string  true  "Organization ID"
// @Param        invitationId  path      string  true  "Invitation ID"
// @Success      204
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization or pending invitation not found"
// @Router       /organizations/{id}/invitations/{invitationId} [delete]
func (h OrganizationHttpHandler) RevokeInvitation(c *gin.Context) {
	if _, err := h.organizationClient.RevokeInvitation(c, &users.InvitationActionRequest{
		OrganizationId: c.Param("id"),
		InvitationId:   c.Param("invitationId"),
		ActorId:        c.GetString("userId"),
	}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      List pending invitations for the current user's email
// @Tags         Invitations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=[]dto.InvitationResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /invitations/me [get]
func (h OrganizationHttpHandler) ListMyInvitations(c *gin.Context) {
	res, err := h.organizationClient.ListMyInvitations(c, &users.ListMyInvitationsRequest{Email: c.GetString("email")})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewInvitationListResponse(res)})
}

// @Summary      Accept an invitation
// @Description  Joins the organization with the invited role. The invitation must be addressed to the current user's email. The user gets the ORGANIZER role (visible after the next token refresh).
// @Tags         Invitations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Invitation ID"
// @Success      200  {object}  object{data=dto.MemberResponse}
// @Failure      400  {object}  object{error=string}  "Invitation expired, revoked, or already answered"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Invitation not found"
// @Failure      409  {object}  object{error=string}  "Already a member"
// @Router       /invitations/{id}/accept [post]
func (h OrganizationHttpHandler) AcceptInvitation(c *gin.Context) {
	res, err := h.organizationClient.AcceptInvitation(c, &users.InvitationActionRequest{
		InvitationId: c.Param("id"),
		ActorId:      c.GetString("userId"),
		ActorEmail:   c.GetString("email"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewMemberResponse(res)})
}

// @Summary      Decline an invitation
// @Tags         Invitations
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Invitation ID"
// @Success      204
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Pending invitation not found"
// @Router       /invitations/{id}/decline [post]
func (h OrganizationHttpHandler) DeclineInvitation(c *gin.Context) {
	if _, err := h.organizationClient.DeclineInvitation(c, &users.InvitationActionRequest{
		InvitationId: c.Param("id"),
		ActorEmail:   c.GetString("email"),
	}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      List all organizations (admin)
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        search    query     string  false  "Match against organization name or email"
// @Param        page      query     int     false  "Page number (default 1)"
// @Param        pageSize  query     int     false  "Page size (default 20, max 100)"
// @Success      200  {object}  object{data=dto.OrganizationPageResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Router       /admin/organizations [get]
func (h OrganizationHttpHandler) List(c *gin.Context) {
	var query dto.ListOrganizationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.organizationClient.ListOrganizations(c, &users.ListOrganizationsRequest{
		Search:   query.Search,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewOrganizationPageResponse(res)})
}

// @Summary      Remove a member, or leave an organization
// @Description  Owners can remove anyone; any member can remove themselves. The last owner cannot be removed. A user who leaves their last organization goes back to the USER role.
// @Tags         Organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Organization ID"
// @Param        userId  path      string  true  "User ID of the member"
// @Success      204
// @Failure      400  {object}  object{error=string}  "Would remove the last owner"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an owner"
// @Failure      404  {object}  object{error=string}  "Organization or member not found"
// @Router       /organizations/{id}/members/{userId} [delete]
func (h OrganizationHttpHandler) RemoveMember(c *gin.Context) {
	if _, err := h.organizationClient.RemoveMember(c, &users.RemoveMemberRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		UserId:         c.Param("userId"),
	}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
