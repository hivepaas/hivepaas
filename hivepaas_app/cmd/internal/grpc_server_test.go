package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
)

var streamInfo = &grpc.StreamServerInfo{FullMethod: "/agent.FileService/FileWrite"}

// A panic in a stream call ends that call with an error and is logged: it does
// not take the agent down with every other call running on it.
func TestAPanicInAStreamCallEndsTheCallNotTheAgent(t *testing.T) {
	logger := &mocks.Logger{}
	interceptor := streamLoggingAndRecoveryInterceptor(logger)

	err := interceptor(nil, nil, streamInfo, func(any, grpc.ServerStream) error {
		panic("a nil map in the handler")
	})

	assert.Equal(t, codes.Internal, status.Code(err))
	if assert.Len(t, logger.Errors, 1) {
		assert.Contains(t, logger.Errors[0], "/agent.FileService/FileWrite")
		assert.Contains(t, logger.Errors[0], "a nil map in the handler")
	}
}

// A stream call's own error crosses as it is, and its result when it has none.
func TestAStreamCallsErrorCrossesAsItIs(t *testing.T) {
	interceptor := streamLoggingAndRecoveryInterceptor(&mocks.Logger{})

	err := interceptor(nil, nil, streamInfo, func(any, grpc.ServerStream) error {
		return hperrors.ToGRPCError(hperrors.NewNotFound("File"))
	})
	assert.Equal(t, codes.NotFound, status.Code(err))

	err = interceptor(nil, nil, streamInfo, func(any, grpc.ServerStream) error {
		return hperrors.NewNotFound("File")
	})
	assert.Equal(t, codes.NotFound, status.Code(err))

	err = interceptor(nil, nil, streamInfo, func(any, grpc.ServerStream) error { return nil })
	assert.NoError(t, err)

	err = interceptor(nil, nil, streamInfo, func(any, grpc.ServerStream) error { return errors.New("plain") })
	assert.Error(t, err)
}
