package apperrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserNotFound       = status.Error(codes.NotFound, "User not found.")
	ErrUserFailedUpdate   = status.Error(codes.Internal, "Failed to update user.")
	ErrInvalidAvatar      = status.Error(codes.InvalidArgument, "Avatar must be a JPEG, PNG or WebP image up to 2 MB.")
	ErrAvatarUploadFailed = status.Error(codes.Internal, "Failed to upload the avatar.")
)

var (
	ErrApplicationNotFound        = status.Error(codes.NotFound, "Organizer application not found.")
	ErrApplicationAlreadyPending  = status.Error(codes.AlreadyExists, "You already have an organizer application under review.")
	ErrApplicationAlreadyReviewed = status.Error(codes.FailedPrecondition, "This application has already been reviewed.")
)

var (
	ErrOrganizationNotFound = status.Error(codes.NotFound, "Organization not found.")
	ErrNotOrganizationOwner = status.Error(codes.PermissionDenied, "Only organization owners can do this.")
	ErrAlreadyMember        = status.Error(codes.AlreadyExists, "This user is already a member of the organization.")
	ErrMemberNotFound       = status.Error(codes.NotFound, "This user is not a member of the organization.")
	ErrLastOwner            = status.Error(codes.FailedPrecondition, "An organization must keep at least one owner.")
	ErrInvalidMemberRole    = status.Error(codes.InvalidArgument, "Member role must be OWNER or STAFF.")
	ErrInvalidMemberStatus  = status.Error(codes.InvalidArgument, "Member status must be ACTIVE or INACTIVE.")
	ErrChangeOwnStatus      = status.Error(codes.FailedPrecondition, "You can't change your own status.")
	ErrLastActiveOwner      = status.Error(codes.FailedPrecondition, "An organization must keep at least one active owner.")
	ErrInviteeNotRegistered = status.Error(codes.FailedPrecondition, "No debyut account uses this email. Ask them to sign up first.")
	ErrInviteeIsAdmin       = status.Error(codes.FailedPrecondition, "Admin accounts can't join organizations.")

	ErrOrganizationNameTaken  = status.Error(codes.AlreadyExists, "An organization with this name already exists.")
	ErrOrganizationEmailTaken = status.Error(codes.AlreadyExists, "An organization already uses this email.")
)

var (
	ErrOrganizationEmailNotVerified = status.Error(codes.FailedPrecondition, "Your organization email verification has expired. Please verify the email again.")
	ErrInvitationNotFound           = status.Error(codes.NotFound, "Invitation not found.")
	ErrInvitationNotPending         = status.Error(codes.FailedPrecondition, "This invitation is no longer valid.")
	ErrInvitationAlreadyPending     = status.Error(codes.AlreadyExists, "This email already has a pending invitation to the organization.")
)
