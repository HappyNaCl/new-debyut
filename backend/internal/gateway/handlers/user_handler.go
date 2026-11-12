package handlers

import (
	"net/http"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/common/grpcerror"
	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/users"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHttpHandler struct {
	userClient users.UserServiceClient
}

func NewUserHttpHandler(
	userClient users.UserServiceClient,
) *UserHttpHandler {
	return &UserHttpHandler{
		userClient: userClient,
	}
}

// @Summary      Get the current user's profile
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{data=dto.MeResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /me [get]
func (u UserHttpHandler) GetMe(c *gin.Context) {
	userId := c.GetString("userId")
	role := c.GetString("role")

	me, err := u.userClient.GetMe(c.Request.Context(), &users.GetMeRequest{Id: userId})
	if err != nil {
		c.JSON(grpcerror.HTTPStatus(err), gin.H{"error": grpcerror.Message(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewMeResponse(me, role)})
}

// @Summary      Update the current user's profile
// @Description  Replaces the profile fields, including the newsletter opt-in. Users must stay 18+.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateMeRequest  true  "Profile fields"
// @Success      200  {object}  object{data=dto.MeResponse}
// @Failure      400  {object}  object{error=object}  "Validation error or under-age"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /me [put]
func (u UserHttpHandler) UpdateMe(c *gin.Context) {
	var req dto.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"dateOfBirth": "is not a valid date"}})
		return
	}

	me, err := u.userClient.UpdateMe(c.Request.Context(), &users.UpdateMeRequest{
		Id:                  c.GetString("userId"),
		FirstName:           req.FirstName,
		Surname:             req.Surname,
		DateOfBirth:         timestamppb.New(dateOfBirth),
		PhoneNumber:         req.PhoneNumber,
		Gender:              req.Gender,
		SubscribeNewsletter: req.SubscribeNewsletter,
	})
	if err != nil {
		c.JSON(grpcerror.HTTPStatus(err), gin.H{"error": grpcerror.Message(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewMeResponse(me, c.GetString("role"))})
}

// @Summary      Upload the current user's avatar
// @Description  Replaces the avatar with a JPEG, PNG or WebP image up to 2 MB.
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Avatar image"
// @Success      200  {object}  object{data=dto.MeResponse}
// @Failure      400  {object}  object{error=object}  "Missing, too large, or unsupported image"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Router       /me/avatar [put]
func (u UserHttpHandler) UpdateAvatar(c *gin.Context) {
	content, ok := readAvatar(c)
	if !ok {
		return
	}

	me, err := u.userClient.UpdateAvatar(c.Request.Context(), &users.UpdateAvatarRequest{
		Id:      c.GetString("userId"),
		Content: content,
	})
	if err != nil {
		c.JSON(grpcerror.HTTPStatus(err), gin.H{"error": grpcerror.Message(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewMeResponse(me, c.GetString("role"))})
}
