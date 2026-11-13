package grpcerror

import (
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func Message(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}

// FieldViolations returns the field errors attached to a gRPC status as
// errdetails.BadRequest, keyed by field. It returns nil when there are none.
func FieldViolations(err error) map[string]string {
	s, ok := status.FromError(err)
	if !ok {
		return nil
	}
	var fields map[string]string
	for _, detail := range s.Details() {
		if badRequest, ok := detail.(*errdetails.BadRequest); ok {
			for _, violation := range badRequest.GetFieldViolations() {
				if fields == nil {
					fields = map[string]string{}
				}
				fields[violation.GetField()] = violation.GetDescription()
			}
		}
	}
	return fields
}
