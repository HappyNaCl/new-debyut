package handlers

import (
	"net/http"

	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/organizers"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

type OrganizerApplicationHttpHandler struct {
	applicationClient users.OrganizerApplicationServiceClient
}

func NewOrganizerApplicationHttpHandler(applicationClient users.OrganizerApplicationServiceClient) *OrganizerApplicationHttpHandler {
	return &OrganizerApplicationHttpHandler{applicationClient: applicationClient}
}

// @Summary      Verify the organization email (step 1 — email OTP)
// @Description  Emails a one-time password to the organization's email. Calling it again resends the code (1-minute cooldown). The code expires after 5 minutes or 5 wrong attempts.
// @Tags         Organizer Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.OrganizationEmailRequest  true  "Organization email"
// @Success      202
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Admins cannot apply"
// @Failure      429  {object}  object{error=string}  "OTP requested too recently"
// @Router       /organizer-applications/email [post]
func (h OrganizerApplicationHttpHandler) RequestEmailOtp(c *gin.Context) {
	var req dto.OrganizationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	if _, err := h.applicationClient.RequestEmailOtp(c, &users.RequestEmailOtpRequest{
		UserId: c.GetString("userId"),
		Email:  req.Email,
	}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

// @Summary      Verify the organization email OTP (step 2)
// @Description  Exchanges a valid OTP for a verification token (valid 30 minutes) that authorizes submitting the application.
// @Tags         Organizer Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.VerifyOrganizationEmailRequest  true  "Organization email and 6-digit OTP"
// @Success      200  {object}  object{data=dto.VerificationTokenResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, or invalid/expired OTP"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Admins cannot apply"
// @Router       /organizer-applications/email/verify [post]
func (h OrganizerApplicationHttpHandler) VerifyEmailOtp(c *gin.Context) {
	var req dto.VerifyOrganizationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.applicationClient.VerifyEmailOtp(c, &users.VerifyEmailOtpRequest{
		UserId: c.GetString("userId"),
		Email:  req.Email,
		Otp:    req.Otp,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.VerificationTokenResponse{VerificationToken: res.GetVerificationToken()}})
}

// @Summary      Submit the organizer application (step 3 — Work with us)
// @Description  Submits the application for admin review using the verified organization email. Users and existing organizers can apply (an organizer applies to create another organization). A user may have one pending application at a time and can apply again after a decline.
// @Tags         Organizer Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.SubmitApplicationRequest  true  "Verification token and organization details"
// @Success      201  {object}  object{data=dto.ApplicationResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, or email verification expired"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Admins cannot apply"
// @Failure      409  {object}  object{error=string}  "An application is already pending"
// @Router       /organizer-applications [post]
func (h OrganizerApplicationHttpHandler) Submit(c *gin.Context) {
	var req dto.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.applicationClient.SubmitApplication(c, &users.SubmitApplicationRequest{
		UserId:            c.GetString("userId"),
		VerificationToken: req.VerificationToken,
		OrganizationName:  req.OrganizationName,
		Description:       req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.NewApplicationResponse(res)})
}

// @Summary      List the current user's organizer applications
// @Description  Newest first, including declined ones.
// @Tags         Organizer Applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=dto.ApplicationListResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /organizer-applications/me [get]
func (h OrganizerApplicationHttpHandler) ListMine(c *gin.Context) {
	res, err := h.applicationClient.ListMyApplications(c, &users.ListMyApplicationsRequest{UserId: c.GetString("userId")})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewApplicationListResponse(res)})
}

// @Summary      List organizer applications (admin)
// @Tags         Organizer Applications
// @Produce      json
// @Security     BearerAuth
// @Param        status    query     string  false  "Filter by status"  Enums(PENDING, APPROVED, DECLINED)
// @Param        page      query     int     false  "Page number (default 1)"
// @Param        pageSize  query     int     false  "Page size (default 20, max 100)"
// @Success      200  {object}  object{data=dto.ApplicationListResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Router       /admin/organizer-applications [get]
func (h OrganizerApplicationHttpHandler) List(c *gin.Context) {
	var query dto.ListApplicationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.applicationClient.ListApplications(c, &users.ListApplicationsRequest{
		Status:   users.ApplicationStatus(users.ApplicationStatus_value[query.Status]),
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewApplicationListResponse(res)})
}

// @Summary      Approve an organizer application (admin)
// @Description  Marks the application approved, creates the organization with the applicant as its OWNER, and gives the applicant the ORGANIZER role. The new role appears in the applicant's next access token (after a refresh).
// @Tags         Organizer Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Application ID"
// @Success      200  {object}  object{data=dto.ApplicationResponse}
// @Failure      400  {object}  object{error=string}  "Application already reviewed"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Failure      404  {object}  object{error=string}  "Application not found"
// @Router       /admin/organizer-applications/{id}/approve [post]
func (h OrganizerApplicationHttpHandler) Approve(c *gin.Context) {
	res, err := h.applicationClient.ApproveApplication(c, &users.ReviewApplicationRequest{
		ApplicationId: c.Param("id"),
		ReviewerId:    c.GetString("userId"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewApplicationResponse(res)})
}

// @Summary      Decline an organizer application (admin)
// @Description  Marks the application declined. The applicant may submit a new application afterwards.
// @Tags         Organizer Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Application ID"
// @Success      200  {object}  object{data=dto.ApplicationResponse}
// @Failure      400  {object}  object{error=string}  "Application already reviewed"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Failure      404  {object}  object{error=string}  "Application not found"
// @Router       /admin/organizer-applications/{id}/decline [post]
func (h OrganizerApplicationHttpHandler) Decline(c *gin.Context) {
	res, err := h.applicationClient.DeclineApplication(c, &users.ReviewApplicationRequest{
		ApplicationId: c.Param("id"),
		ReviewerId:    c.GetString("userId"),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewApplicationResponse(res)})
}
