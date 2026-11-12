package dto

type CreateInvitationRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
	Role  string `json:"role"  binding:"required,oneof=OWNER STAFF"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name"        binding:"required,max=150"`
	Description string `json:"description" binding:"required,max=2000"`
}

type ListOrganizationsQuery struct {
	Search   string `form:"search"   binding:"omitempty,max=150"`
	Page     int32  `form:"page"     binding:"omitempty,min=1"`
	PageSize int32  `form:"pageSize" binding:"omitempty,min=1,max=100"`
}
