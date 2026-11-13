package handlers

import (
	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/common/grpcerror"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// respondFieldError reports a known service error as a field error
// ({"error": {"<field>": "<message>"}}, with the usual HTTP status for the
// error) so forms can show it next to the right input. Any other error is
// reported as usual.
func respondFieldError(c *gin.Context, err error, fields map[error]string) {
	for known, field := range fields {
		if field != "" && sameStatus(err, known) {
			c.JSON(grpcerror.HTTPStatus(err), gin.H{"error": gin.H{field: grpcerror.Message(err)}})
			return
		}
	}
	respondError(c, err)
}

// respondOrganizationError reports organization name/email conflicts as field
// errors under the given field names (empty to skip one).
func respondOrganizationError(c *gin.Context, err error, nameField, emailField string) {
	respondFieldError(c, err, map[error]string{
		apperrors.ErrOrganizationNameTaken:  nameField,
		apperrors.ErrOrganizationEmailTaken: emailField,
	})
}

func sameStatus(err, target error) bool {
	got, ok := status.FromError(err)
	if !ok {
		return false
	}
	want := status.Convert(target)
	return got.Code() == want.Code() && got.Message() == want.Message()
}
