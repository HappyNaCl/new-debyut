package dto

type ListApplicationsQuery struct {
	Status   string `form:"status"   binding:"omitempty,oneof=PENDING APPROVED DECLINED"`
	Page     int32  `form:"page"     binding:"omitempty,min=1"`
	PageSize int32  `form:"pageSize" binding:"omitempty,min=1,max=100"`
}
