package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
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

// agentConfig installs a config carrying the given agent secret and dev mode for
// the length of the test: validateAuthToken reads the config in force, the way
// every other per-operation read in this codebase does.
func agentConfig(t *testing.T, secretToken string, devMode bool) {
	t.Helper()
	previous := config.Current()
	cfg := &config.Config{Agent: config.Agent{SecretToken: secretToken}}
	cfg.DevMode.Enabled = devMode
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })
}

// ctxWithToken is an incoming call carrying the token the control plane sends.
func ctxWithToken(token string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", token))
}

// serverStream is a grpc.ServerStream that has nothing but a context: the auth
// interceptor reads that and nothing else of it.
type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s serverStream) Context() context.Context { return s.ctx }

// The configured secret admits a call, and nothing else does.
func TestTheConfiguredSecretAdmitsACallAndNothingElseDoes(t *testing.T) {
	agentConfig(t, "the-secret", false)

	assert.NoError(t, validateAuthToken(ctxWithToken("the-secret")))

	for _, token := range []string{"wrong", "", "the-secret "} {
		err := validateAuthToken(ctxWithToken(token))
		assert.Equal(t, codes.Unauthenticated, status.Code(err), "token %q", token)
	}

	noAuthKey := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "value"))
	assert.Equal(t, codes.Unauthenticated, status.Code(validateAuthToken(noAuthKey)))
	assert.Equal(t, codes.Unauthenticated, status.Code(validateAuthToken(context.Background())))
}

// The agent read AGENT_SECRET_TOKEN from the environment while every deployment
// sets HP_AGENT_SECRET_TOKEN, so it found nothing and passed every call through
// unauthenticated while the control plane dutifully sent the token. A value in
// the environment under a name of the agent's own is not a configured secret.
func TestASecretInTheEnvironmentIsNotTheAgentsSecret(t *testing.T) {
	t.Setenv("AGENT_SECRET_TOKEN", "from-the-environment")
	agentConfig(t, "", false)

	err := validateAuthToken(ctxWithToken("from-the-environment"))
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

// A secret that is not configured is a misconfiguration, not permission to skip
// the check - outside dev and local, where the token is optional by design.
func TestAnAgentWithNoSecretRefusesEveryCallButInDevMode(t *testing.T) {
	agentConfig(t, "", false)
	assert.Equal(t, codes.Unauthenticated, status.Code(validateAuthToken(ctxWithToken("anything"))))
	assert.Equal(t, codes.Unauthenticated, status.Code(validateAuthToken(context.Background())))

	agentConfig(t, "", true)
	assert.NoError(t, validateAuthToken(context.Background()))
}

// The config is loaded long before the server starts. If it ever is not, the
// call is refused rather than admitted.
func TestACallIsRefusedWhileTheConfigIsNotLoaded(t *testing.T) {
	previous := config.Current()
	config.SetCurrent(nil)
	t.Cleanup(func() { config.SetCurrent(previous) })

	err := validateAuthToken(ctxWithToken("anything"))
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

// The interceptors are where the check takes effect: a refused call must not
// reach the handler, on a unary call and on a stream alike.
func TestARefusedCallNeverReachesTheHandler(t *testing.T) {
	agentConfig(t, "the-secret", false)

	var reached bool
	unaryHandler := func(context.Context, any) (any, error) { reached = true; return "answer", nil }
	streamHandler := func(any, grpc.ServerStream) error { reached = true; return nil }

	_, err := unaryAuthInterceptor()(ctxWithToken("wrong"), nil, nil, unaryHandler)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.False(t, reached)

	err = streamAuthInterceptor()(nil, serverStream{ctx: ctxWithToken("wrong")}, streamInfo, streamHandler)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.False(t, reached)

	resp, err := unaryAuthInterceptor()(ctxWithToken("the-secret"), nil, nil, unaryHandler)
	assert.NoError(t, err)
	assert.Equal(t, "answer", resp)
	assert.True(t, reached)

	reached = false
	err = streamAuthInterceptor()(nil, serverStream{ctx: ctxWithToken("the-secret")}, streamInfo, streamHandler)
	assert.NoError(t, err)
	assert.True(t, reached)
}
