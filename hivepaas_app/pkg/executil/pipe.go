package executil

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
)

// execBufferMaxSize limits the maximum capacity of a buffer to be reused in the pool.
// Any buffer larger than 64KB will be discarded to let the Garbage Collector reclaim its memory.
const execBufferMaxSize = 64 * 1024 // 64 KB

// execBufferPool is a thread-safe pool used to reuse bytes.Buffer instances,
// which reduces memory allocations and GC pressure under high execution loads.
var execBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// getExecBuffer retrieves a clean buffer from the pool.
func getExecBuffer() *bytes.Buffer {
	buf := execBufferPool.Get().(*bytes.Buffer) //nolint
	buf.Reset()
	return buf
}

// putExecBuffer returns a buffer to the pool if its capacity is within limits.
// Discards oversized buffers to prevent memory bloating (Buffer Bloat).
func putExecBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if buf.Cap() <= execBufferMaxSize {
		execBufferPool.Put(buf)
	}
}

// safeWriter is a thread-safe wrapper around bytes.Buffer to prevent data races
// when multiple concurrent commands write stdout/stderr.
type safeWriter struct {
	mu  sync.Mutex
	buf *bytes.Buffer
}

func (w *safeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	n, err = w.buf.Write(p)
	w.mu.Unlock()
	return
}

func (w *safeWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// RunPipeline executes a series of commands in a pipeline (cmd1 | cmd2 | ... | cmdN).
// It connects the standard output of each command to the standard input of the next.
// It combines the stdout of the final command and the stderr of all commands into a single output,
// similar to cmd.CombinedOutput().
func RunPipeline(cmds ...*exec.Cmd) (string, error) {
	numCmds := len(cmds)
	if numCmds == 0 {
		return "", nil
	}

	lastCmd := cmds[numCmds-1]
	precedingCmds := cmds[:numCmds-1]

	// 1. Pipe the stdout of each command to the stdin of the next command
	for i, nextCmd := range cmds[1:] {
		prevCmd := cmds[i]
		stdout, err := prevCmd.StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create stdout pipe for command %d (%s): %w",
				i, prevCmd.Path, err)
		}
		nextCmd.Stdin = stdout
	}

	// 2. Retrieve a pooled buffer for the combined output and defer its return
	outBuf := getExecBuffer()
	defer putExecBuffer(outBuf)

	// Wrap in a thread-safe writer to prevent data races when multiple commands
	// write stdout/stderr concurrently.
	writer := &safeWriter{buf: outBuf}

	// 3. Connect final stdout and all stderrs to the combined writer
	lastCmd.Stdout = writer
	for _, cmd := range cmds {
		cmd.Stderr = writer
	}

	// 4. Start all preceding commands asynchronously
	// Using Start() allows the commands to execute concurrently and stream data through pipes
	var startedCmds []*exec.Cmd
	for i, cmd := range precedingCmds {
		if err := cmd.Start(); err != nil {
			for _, started := range startedCmds {
				_ = started.Wait()
			}
			return writer.String(), fmt.Errorf("failed to start command %d (%s): %w",
				i, cmd.Path, err)
		}
		startedCmds = append(startedCmds, cmd)
	}

	// 5. Execute the final command synchronously (Run starts the command and waits for it to finish)
	lastErr := lastCmd.Run()

	// 6. Always wait for all preceding commands to release their resources and flush I/O
	var waitErr error
	for i, cmd := range precedingCmds {
		if err := cmd.Wait(); err != nil && waitErr == nil {
			waitErr = fmt.Errorf("command %d (%s) failed during wait: %w",
				i, cmd.Path, err)
		}
	}

	output := writer.String()
	if lastErr != nil {
		return output, fmt.Errorf("failed to run last command (%s): %w",
			lastCmd.Path, lastErr)
	}
	if waitErr != nil {
		return output, waitErr
	}

	return output, nil
}
