package apperrors

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCategoryNotFound  = status.Error(codes.NotFound, "Category not found.")
	ErrCategorySlugTaken = status.Error(codes.AlreadyExists, "Another category already uses this slug.")
	ErrCategoryNameTaken = status.Error(codes.AlreadyExists, "Another category already uses this name.")
	ErrInvalidCategory   = status.Error(codes.InvalidArgument, "Category slug and icon must be lowercase letters, numbers, and hyphens.")
)

var (
	ErrCategoryInUse      = status.Error(codes.FailedPrecondition, "This category is used by events and can't be deleted.")
	ErrEventNotFound      = status.Error(codes.NotFound, "Event not found.")
	ErrEventNotEditable   = status.Error(codes.FailedPrecondition, "Published events can't be edited yet.")
	ErrInvalidEventImage  = status.Error(codes.InvalidArgument, "Image must be a JPEG, PNG or WebP file within the size limit.")
	ErrInvalidImageKind   = status.Error(codes.InvalidArgument, "Image kind must be BANNER, COVER or GALLERY.")
	ErrEventImageUpload   = status.Error(codes.Internal, "Failed to upload the image.")
	ErrInvalidEventFilter = status.Error(codes.InvalidArgument, "Filter must be DRAFT, ACTIVE or PAST.")
)

// EventValidationError lists what is wrong with an event, keyed by field
// (e.g. "name", "tickets.0.price"). Publishing checks every rule; drafts only
// check the shape of what was filled in.
func EventValidationError(fields map[string]string) error {
	st := status.New(codes.InvalidArgument, "Please fix the highlighted fields.")
	details := &errdetails.BadRequest{}
	for field, message := range fields {
		details.FieldViolations = append(details.FieldViolations, &errdetails.BadRequest_FieldViolation{Field: field, Description: message})
	}
	if withDetails, err := st.WithDetails(details); err == nil {
		return withDetails.Err()
	}
	return st.Err()
}
