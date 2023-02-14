package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFieldViolation(t *testing.T) {
	errDetails := FieldViolation("testField", fmt.Errorf("testMsg"))
	require.Equal(t, errDetails.Field, "testField")
	require.Equal(t, errDetails.Description, "testMsg")

	err := InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{errDetails})
	e, _ := status.FromError(err)
	require.Equal(t, codes.InvalidArgument, e.Code())
}
