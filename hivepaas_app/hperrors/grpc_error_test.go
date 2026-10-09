package hperrors

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToGRPCErrorKeepsEveryErrorAnError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"an error of a known base", NewNotFound("App"), codes.NotFound},
		{"an error wrapped from outside", Wrap(errors.New("no Dockerfile")), codes.Unknown},
		{"a panic", NewPanic("boom"), codes.Unknown},
		{"an error of no kind of ours", errors.New("plain"), codes.Internal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToGRPCError(tc.err)
			if got == nil {
				t.Fatalf("ToGRPCError(%v) = nil: the call would end as if it had succeeded", tc.err)
			}
			if code := status.Code(got); code != tc.code {
				t.Errorf("ToGRPCError(%v) code = %v; want %v", tc.err, code, tc.code)
			}
		})
	}
	if ToGRPCError(nil) != nil {
		t.Error("ToGRPCError(nil) is an error")
	}
}
