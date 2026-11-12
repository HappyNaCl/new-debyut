package apperrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCategoryNotFound  = status.Error(codes.NotFound, "Category not found.")
	ErrCategorySlugTaken = status.Error(codes.AlreadyExists, "Another category already uses this slug.")
	ErrCategoryNameTaken = status.Error(codes.AlreadyExists, "Another category already uses this name.")
	ErrInvalidCategory   = status.Error(codes.InvalidArgument, "Category slug and icon must be lowercase letters, numbers, and hyphens.")
)
