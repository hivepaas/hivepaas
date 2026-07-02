package ssh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"
	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

const (
	defaultConnectTimeout = 10 * time.Second
)

var (
	ErrExecutionTimeout = errors.New("execution timeout")
)

type CommandInput struct {
	Host           string
	Port           int
	User           string
	PrivateKey     string
	Passphrase     string
	ConnectTimeout time.Duration // default: 10s
	Command        string
}

func Execute(ctx context.Context, input *CommandInput) (output string, err error) {
	var signer ssh.Signer
	if input.Passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(input.PrivateKey), []byte(input.Passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(input.PrivateKey))
	}
	if err != nil {
		return "", apperrors.New(err).WithMsgLog("failed to parse private key")
	}

	config := &ssh.ClientConfig{
		User: input.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: gofn.If(input.ConnectTimeout > 0, input.ConnectTimeout, defaultConnectTimeout), //nolint
		// TODO: handle host key validation for higher security
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	// Connect to the remote server
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", input.Host, input.Port), config)
	if err != nil {
		return "", apperrors.New(err).WithMsgLog("failed to dial ssh")
	}
	defer client.Close()

	// Create a new session
	session, err := client.NewSession()
	if err != nil {
		return "", apperrors.New(err).WithMsgLog("failed to create session")
	}
	defer session.Close()

	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		// Run the command and get combined output
		outBytes, err := session.CombinedOutput(input.Command)
		if err != nil {
			return "", apperrors.New(err).WithMsgLog("failed to execute command")
		}
		return string(outBytes), nil
	}

	type outputStruct struct {
		Error  error
		Output string
	}

	done := make(chan *outputStruct, 1)
	go func() {
		outBytes, err := session.CombinedOutput(input.Command)
		done <- &outputStruct{
			Error:  err,
			Output: string(outBytes),
		}
	}()

	select {
	case <-ctx.Done():
		// Timeout
		// You might want to send a signal to the remote process to terminate it gracefully,
		// though this can be complex with standard SSH sessions.
		return "", apperrors.New(ErrExecutionTimeout)
	case outputData := <-done:
		if outputData.Error != nil {
			return "", apperrors.New(outputData.Error).WithMsgLog("failed to execute command")
		} else {
			return outputData.Output, nil
		}
	}
}
