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

// RunPipeline executes a series of commands in a pipeline (cmd1 | cmd2 | ... | cmdN).
// It connects the standard output of each command to the standard input of the next.
// It combines the stdout of the final command and the stderr of all commands into a single output,
// similar to cmd.CombinedOutput().
func RunPipeline(cmds ...*exec.Cmd) (string, error) {
	if len(cmds) == 0 {
		return "", nil
	}

	// 1. Pipe the stdout of each command to the stdin of the next command
	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create stdout pipe for command %d (%s): %w",
				i, cmds[i].Path, err)
		}
		cmds[i+1].Stdin = stdout
	}

	// 2. Retrieve a pooled buffer for the combined output and defer its return
	outBuf := getExecBuffer()
	defer putExecBuffer(outBuf)

	// Wrap in a thread-safe writer to prevent data races when multiple commands
	// write stdout/stderr concurrently.
	writer := &safeWriter{buf: outBuf}

	// 3. Connect final stdout and all stderrs to the combined writer
	cmds[len(cmds)-1].Stdout = writer
	for i := range cmds {
		cmds[i].Stderr = writer
	}

	// 4. Start all preceding commands asynchronously
	// Using Start() allows the commands to execute concurrently and stream data through pipes
	for i := 0; i < len(cmds)-1; i++ {
		if err := cmds[i].Start(); err != nil {
			return outBuf.String(), fmt.Errorf("failed to start command %d (%s): %w",
				i, cmds[i].Path, err)
		}
	}

	// 5. Execute the final command synchronously (Run starts the command and waits for it to finish)
	if err := cmds[len(cmds)-1].Run(); err != nil {
		return outBuf.String(), fmt.Errorf("failed to run last command (%s): %w",
			cmds[len(cmds)-1].Path, err)
	}

	// 6. Wait for all preceding commands to release their resources and exit
	for i := 0; i < len(cmds)-1; i++ {
		if err := cmds[i].Wait(); err != nil {
			return outBuf.String(), fmt.Errorf("command %d (%s) failed during wait: %w",
				i, cmds[i].Path, err)
		}
	}
	return outBuf.String(), nil
}
