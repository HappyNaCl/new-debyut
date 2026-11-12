package dto

type CategoryRequest struct {
	Name string `json:"name" binding:"required,max=100"`
	Slug string `json:"slug" binding:"omitempty,max=64"`
	Icon string `json:"icon" binding:"required,max=64"`
}
