package nodeexecservice

import "io"

type CommandExecReq struct {
	NodeID    string
	NodeLabel string
	*CommandExecOpts
}

type CommandExecOpts struct {
	Command    []string
	Env        []string
	WorkingDir string
	Stdout     io.Writer
	Stderr     io.Writer // if nil, errors will go through Stdout, use io.Discard to discard
}

type CommandExecResp struct {
	ExitCode int32
}
