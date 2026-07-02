package containerservice

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	"github.com/moby/moby/client"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/services/docker"
)

// TerminalSize holds the dimensions of a console terminal.
type TerminalSize struct {
	Width  uint32
	Height uint32
}

// ContainerExecConfig represents the initial configuration for command execution in a container.
type ContainerExecConfig struct {
	ContainerID string
	Cmd         []string
	Env         []string
	WorkingDir  string
	Tty         bool
	ConsoleSize *TerminalSize
}

// ContainerExecReq represents incoming/outgoing stream events for execution.
type ContainerExecReq struct {
	Config *ContainerExecConfig
	Stdin  []byte
	Resize *TerminalSize
}

// ContainerExecResp represents outgoing stream events containing terminal output or exit code.
type ContainerExecResp struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode *int32
}

type ContainerExecStream struct {
	stream     agentproto.ContainerService_ContainerExecClient
	cancelFunc context.CancelFunc
	grpcConn   *grpc.ClientConn

	// Buffering output from gRPC stream
	readBuf   bytes.Buffer
	readMutex sync.Mutex
	readErr   error
	cond      *sync.Cond

	// Exit code captured from stream
	exitCode int32
	hasExit  bool
}

func (s *ContainerExecStream) start() {
	s.cond = sync.NewCond(&s.readMutex)

	// Start reading from stream in background
	go s.readStreamLoop()
}

// Context returns the context associated with the stream.
func (s *ContainerExecStream) Context() context.Context {
	return s.stream.Context()
}

// Send sends a command execution request stream event.
func (s *ContainerExecStream) Send(req *ContainerExecReq) error {
	protoReq := &agentproto.ContainerExecReq{}
	switch {
	case req.Config != nil:
		var size *agentproto.TerminalSize
		if req.Config.ConsoleSize != nil {
			size = &agentproto.TerminalSize{
				Width:  req.Config.ConsoleSize.Width,
				Height: req.Config.ConsoleSize.Height,
			}
		}
		protoReq.Value = &agentproto.ContainerExecReq_Config{
			Config: &agentproto.ContainerExecConfig{
				ContainerId: req.Config.ContainerID,
				Cmd:         req.Config.Cmd,
				Env:         req.Config.Env,
				WorkingDir:  req.Config.WorkingDir,
				Tty:         req.Config.Tty,
				ConsoleSize: size,
			},
		}
	case req.Stdin != nil:
		protoReq.Value = &agentproto.ContainerExecReq_Stdin{
			Stdin: req.Stdin,
		}
	case req.Resize != nil:
		protoReq.Value = &agentproto.ContainerExecReq_Resize{
			Resize: &agentproto.TerminalSize{
				Width:  req.Resize.Width,
				Height: req.Resize.Height,
			},
		}
	}
	return s.stream.Send(protoReq) //nolint:wrapcheck
}

func (s *ContainerExecStream) SendExecCreate(containerID string, option docker.ExecCreateOption) error {
	var execOpts client.ExecCreateOptions
	if option != nil {
		option(&execOpts)
	}

	var consoleSize *TerminalSize
	if execOpts.ConsoleSize.Width > 0 || execOpts.ConsoleSize.Height > 0 {
		consoleSize = &TerminalSize{
			Width:  uint32(execOpts.ConsoleSize.Width),  //nolint:gosec
			Height: uint32(execOpts.ConsoleSize.Height), //nolint:gosec
		}
	}

	return s.Send(&ContainerExecReq{
		Config: &ContainerExecConfig{
			ContainerID: containerID,
			Cmd:         execOpts.Cmd,
			Env:         execOpts.Env,
			WorkingDir:  execOpts.WorkingDir,
			Tty:         execOpts.TTY,
			ConsoleSize: consoleSize,
		},
	})
}

func (s *ContainerExecStream) SendResize(width, height uint) error {
	err := s.Send(&ContainerExecReq{
		Resize: &TerminalSize{
			Width:  uint32(width),  //nolint:gosec
			Height: uint32(height), //nolint:gosec
		},
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

// Recv receives a command execution response stream event.
func (s *ContainerExecStream) Recv() (*ContainerExecResp, error) {
	resp, err := s.stream.Recv()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	out := &ContainerExecResp{}
	if resp.GetStdout() != nil {
		out.Stdout = resp.GetStdout()
	} else if resp.GetStderr() != nil {
		out.Stderr = resp.GetStderr()
	} else if exitCodeVal, ok := resp.GetValue().(*agentproto.ContainerExecResp_ExitCode); ok {
		code := exitCodeVal.ExitCode
		out.ExitCode = &code
	}
	return out, nil
}

// CloseSend closes the send direction of the stream.
func (s *ContainerExecStream) CloseSend() error {
	return s.stream.CloseSend() //nolint:wrapcheck
}

// Helper to create the ExecAttachResult compatible struct
func (s *ContainerExecStream) ToExecAttachResult() *client.ExecAttachResult {
	return &client.ExecAttachResult{
		HijackedResponse: client.HijackedResponse{
			Conn:   s,
			Reader: bufio.NewReader(s),
		},
	}
}

func (s *ContainerExecStream) readStreamLoop() {
	for {
		resp, err := s.Recv()
		s.readMutex.Lock()
		if err != nil {
			s.readErr = err
			s.cond.Broadcast()
			s.readMutex.Unlock()
			return
		}

		switch {
		case resp.Stdout != nil:
			s.readBuf.Write(resp.Stdout)
			s.cond.Broadcast()
		case resp.Stderr != nil:
			s.readBuf.Write(resp.Stderr)
			s.cond.Broadcast()
		case resp.ExitCode != nil:
			s.exitCode = *resp.ExitCode
			s.hasExit = true
			s.cond.Broadcast()
		}
		s.readMutex.Unlock()
	}
}

// Read implements io.Reader
func (s *ContainerExecStream) Read(b []byte) (int, error) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()

	for s.readBuf.Len() == 0 && s.readErr == nil {
		s.cond.Wait()
	}

	if s.readBuf.Len() > 0 {
		n, err := s.readBuf.Read(b)
		return n, apperrors.New(err)
	}

	return 0, s.readErr
}

// Write implements io.Writer (part of net.Conn)
func (s *ContainerExecStream) Write(b []byte) (int, error) {
	// Send stdin bytes via gRPC stream
	err := s.Send(&ContainerExecReq{
		Stdin: b,
	})
	if err != nil {
		return 0, apperrors.New(err)
	}
	return len(b), nil
}

// Close implements io.Closer (part of net.Conn)
func (s *ContainerExecStream) Close() error {
	s.cancelFunc()
	_ = s.stream.CloseSend()
	if s.grpcConn != nil {
		_ = s.grpcConn.Close()
	}
	return nil
}

// Implement dummy methods for net.Conn interface
func (s *ContainerExecStream) LocalAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4zero}
}

func (s *ContainerExecStream) RemoteAddr() net.Addr {
	return &net.IPAddr{IP: net.IPv4zero}
}

func (s *ContainerExecStream) SetDeadline(t time.Time) error {
	return nil
}

func (s *ContainerExecStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *ContainerExecStream) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s *ContainerExecStream) GetExitCode() (int32, bool) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	return s.exitCode, s.hasExit
}
