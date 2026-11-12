package handlers

import (
	"net/http"

	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/categories"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CategoryHttpHandler struct {
	eventClient event.EventServiceClient
}

func NewCategoryHttpHandler(eventClient event.EventServiceClient) *CategoryHttpHandler {
	return &CategoryHttpHandler{eventClient: eventClient}
}

// @Summary      List event categories
// @Description  Public. Sorted by name.
// @Tags         Categories
// @Produce      json
// @Success      200  {object}  object{data=[]dto.CategoryResponse}
// @Router       /categories [get]
func (h CategoryHttpHandler) List(c *gin.Context) {
	res, err := h.eventClient.ListCategories(c, &emptypb.Empty{})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewCategoryListResponse(res)})
}

// @Summary      Create an event category (admin)
// @Description  The slug is generated from the name when omitted. Slug and icon must be lowercase letters, numbers, and hyphens; icon is a key from the frontend's icon set.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CategoryRequest  true  "Category"
// @Success      201  {object}  object{data=dto.CategoryResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Failure      409  {object}  object{error=string}  "Name or slug already used"
// @Router       /admin/categories [post]
func (h CategoryHttpHandler) Create(c *gin.Context) {
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.eventClient.CreateCategory(c, &event.CreateCategoryRequest{Name: req.Name, Slug: req.Slug, Icon: req.Icon})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.NewCategoryResponse(res)})
}

// @Summary      Update an event category (admin)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string               true  "Category ID"
// @Param        request  body      dto.CategoryRequest  true  "Category"
// @Success      200  {object}  object{data=dto.CategoryResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Failure      404  {object}  object{error=string}  "Category not found"
// @Failure      409  {object}  object{error=string}  "Name or slug already used"
// @Router       /admin/categories/{id} [put]
func (h CategoryHttpHandler) Update(c *gin.Context) {
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.eventClient.UpdateCategory(c, &event.UpdateCategoryRequest{Id: c.Param("id"), Name: req.Name, Slug: req.Slug, Icon: req.Icon})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewCategoryResponse(res)})
}

// @Summary      Delete an event category (admin)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Category ID"
// @Success      204
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      403  {object}  object{error=string}  "Not an admin"
// @Failure      404  {object}  object{error=string}  "Category not found"
// @Router       /admin/categories/{id} [delete]
func (h CategoryHttpHandler) Delete(c *gin.Context) {
	if _, err := h.eventClient.DeleteCategory(c, &event.DeleteCategoryRequest{Id: c.Param("id")}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
