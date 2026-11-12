package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/common/grpcerror"
	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/auth"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHttpHandler struct {
	authClient auth.AuthServiceClient
}

const (
	refreshCookieName   = "refreshToken"
	refreshCookiePath   = "/api/auth"
	refreshCookieMaxAge = 7 * 24 * 60 * 60
)

func (a AuthHttpHandler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, token, refreshCookieMaxAge, refreshCookiePath, "", isProduction(), true)
}

func (a AuthHttpHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, refreshCookiePath, "", isProduction(), true)
}

func isProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

func NewAuthHttpHandler(authClient auth.AuthServiceClient) *AuthHttpHandler {
	return &AuthHttpHandler{authClient: authClient}
}

func respondError(c *gin.Context, err error) {
	c.JSON(grpcerror.HTTPStatus(err), gin.H{"error": grpcerror.Message(err)})
}

// @Summary      Start registration (step 1 — email OTP)
// @Description  Emails a one-time password to an address that is not yet registered. Calling it again resends the code (1-minute cooldown). The code expires after 5 minutes or 5 wrong attempts.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.EmailRequest  true  "Email to register"
// @Success      202
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      409  {object}  object{error=string}  "Email already registered"
// @Failure      429  {object}  object{error=string}  "OTP requested too recently"
// @Router       /auth/register [post]
func (a AuthHttpHandler) RequestRegisterOtp(c *gin.Context) {
	var req dto.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	if _, err := a.authClient.RequestRegisterOtp(c, &auth.EmailRequest{Email: req.Email}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

// @Summary      Verify the registration OTP (step 2)
// @Description  Exchanges a valid OTP for a registration token (valid 30 minutes) that authorizes completing the profile.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyOtpRequest  true  "Email and 6-digit OTP"
// @Success      200  {object}  object{data=dto.RegistrationTokenResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, or invalid/expired OTP"
// @Router       /auth/register/verify [post]
func (a AuthHttpHandler) VerifyRegisterOtp(c *gin.Context) {
	var req dto.VerifyOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := a.authClient.VerifyRegisterOtp(c, &auth.VerifyOtpRequest{Email: req.Email, Otp: req.Otp})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.RegistrationTokenResponse{RegistrationToken: res.GetRegistrationToken()}})
}

// @Summary      Complete registration with profile data (step 3)
// @Description  Creates the account for the verified email, stores the profile, and logs the user in. Surname is optional; agreeToTerms must be true; the user must be at least 18.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CompleteRegistrationRequest  true  "Profile data"
// @Success      201  {object}  object{data=dto.AuthResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, under-age, terms not accepted, or registration token expired"
// @Failure      409  {object}  object{error=string}  "Email already registered"
// @Router       /auth/register/complete [post]
func (a AuthHttpHandler) CompleteRegistration(c *gin.Context) {
	var req dto.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"dateOfBirth": "is not a valid date"}})
		return
	}

	res, err := a.authClient.CompleteRegistration(c, &auth.CompleteRegistrationRequest{
		RegistrationToken:   req.RegistrationToken,
		FirstName:           req.FirstName,
		Surname:             req.Surname,
		DateOfBirth:         timestamppb.New(dateOfBirth),
		PhoneNumber:         req.PhoneNumber,
		Gender:              req.Gender,
		AgreeToTerms:        req.AgreeToTerms,
		SubscribeNewsletter: req.SubscribeNewsletter,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	a.setRefreshCookie(c, res.GetRefreshToken())
	c.JSON(http.StatusCreated, gin.H{"data": dto.NewAuthResponse(res)})
}

// @Summary      Start login (step 1 — email OTP)
// @Description  Emails a one-time password to a registered address. Calling it again resends the code (1-minute cooldown).
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.EmailRequest  true  "Registered email"
// @Success      202
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      404  {object}  object{error=string}  "No account with this email"
// @Failure      429  {object}  object{error=string}  "OTP requested too recently"
// @Router       /auth/login [post]
func (a AuthHttpHandler) RequestLoginOtp(c *gin.Context) {
	var req dto.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	if _, err := a.authClient.RequestLoginOtp(c, &auth.EmailRequest{Email: req.Email}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

// @Summary      Verify the login OTP (step 2)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyOtpRequest  true  "Email and 6-digit OTP"
// @Success      200  {object}  object{data=dto.AuthResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, or invalid/expired OTP"
// @Router       /auth/login/verify [post]
func (a AuthHttpHandler) VerifyLoginOtp(c *gin.Context) {
	var req dto.VerifyOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := a.authClient.VerifyLoginOtp(c, &auth.VerifyOtpRequest{Email: req.Email, Otp: req.Otp})
	if err != nil {
		respondError(c, err)
		return
	}

	a.setRefreshCookie(c, res.GetRefreshToken())
	c.JSON(http.StatusOK, gin.H{"data": dto.NewAuthResponse(res)})
}

// @Summary      Rotate the refresh token (from httpOnly cookie) for a new access token
// @Description  Reads the refresh token from the httpOnly refreshToken cookie, rotates it, sets a new cookie, and returns a fresh access token in the body.
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  object{data=dto.TokenResponse}
// @Failure      401  {object}  object{error=string}  "Missing, invalid, expired, or reused refresh token"
// @Router       /auth/refresh [post]
func (a AuthHttpHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}

	res, err := a.authClient.RefreshToken(c, &auth.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		a.clearRefreshCookie(c)
		respondError(c, err)
		return
	}

	a.setRefreshCookie(c, res.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"data": dto.TokenResponse{AccessToken: res.AccessToken}})
}

// @Summary      Revoke all refresh tokens for the current user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{message=string}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /auth/logout [post]
func (a AuthHttpHandler) Logout(c *gin.Context) {
	if _, err := a.authClient.Logout(c, &auth.LogoutRequest{UserId: c.GetString("userId")}); err != nil {
		respondError(c, err)
		return
	}

	a.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
