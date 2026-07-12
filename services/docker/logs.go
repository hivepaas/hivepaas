package docker

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type ScanningLogOptions struct {
	BatchRecvOptions  batchrecvchan.Options
	StdoutWriter      io.Writer
	ParseLogHeader    bool
	ParseLogTimestamp bool
}

type ScanningLogOption func(*ScanningLogOptions)

func WithParseLogHeader(flag bool) ScanningLogOption {
	return func(o *ScanningLogOptions) {
		o.ParseLogHeader = flag
	}
}

func WithParseLogTimestamp(flag bool) ScanningLogOption {
	return func(o *ScanningLogOptions) {
		o.ParseLogTimestamp = flag
	}
}

func WithBatchRecvOptions(recvOpts batchrecvchan.Options) ScanningLogOption {
	return func(o *ScanningLogOptions) {
		o.BatchRecvOptions = recvOpts
	}
}

func WithStdoutWriter(w io.Writer) ScanningLogOption {
	return func(o *ScanningLogOptions) {
		o.StdoutWriter = w
	}
}

func StartScanningLog(
	ctx context.Context,
	reader io.ReadCloser,
	options ...ScanningLogOption,
) (logChan <-chan []*tasklog.LogFrame, closeFunc func() error) {
	opts := &ScanningLogOptions{
		ParseLogHeader: true,
	}
	for _, o := range options {
		o(opts)
	}

	batchChan := batchrecvchan.NewChan[*tasklog.LogFrame](opts.BatchRecvOptions)

	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		context.AfterFunc(ctx, func() { _ = reader.Close() })
	}

	go func() {
		defer func() {
			_ = recover()
			batchChan.Close()
		}()

		// Close logs stream
		defer reader.Close()

		if opts.ParseLogHeader {
			// Docker logs have an 8-byte header for stdout/stderr
			// Ref: https://docs.docker.com/reference/api/engine/version/v1.51/#tag/Container/operation/ContainerAttach
			_ = parseLogs(ctx, reader, batchChan, opts.StdoutWriter, opts.ParseLogTimestamp)
		} else {
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				logFrame := &tasklog.LogFrame{
					Type: tasklog.LogTypeOut,
					Data: reflectutil.UnsafeBytesToStr(scanner.Bytes()),
				}
				if opts.ParseLogTimestamp {
					logFrame.ParseTimestampFromData()
				}

				if ctx.Err() != nil { // context is done
					return
				}

				batchChan.Send(logFrame)
			}
		}
	}()

	return batchChan.Receiver(), func() error { return reader.Close() }
}

// NOTE: This code is copied from `github.com/moby/moby/v2/pkg/stdcopy`
// as we need to add some customization.

const (
	// Stdin represents standard input stream type.
	Stdin byte = 0
	// Stdout represents standard output stream type.
	Stdout byte = 1
	// Stderr represents standard error steam type.
	Stderr byte = 2
	// Systemerr represents errors originating from the system that make it into the multiplexed stream.
	Systemerr byte = 3

	stdWriterPrefixLen = 8
	stdWriterFdIndex   = 0
	stdWriterSizeIndex = 4

	startingBufLen = 32*1024 + stdWriterPrefixLen + 1
)

//nolint:gocognit
func parseLogs(
	ctx context.Context,
	src io.Reader,
	dst *batchrecvchan.Chan[*tasklog.LogFrame],
	dstOfStdout io.Writer,
	parseTimestamp bool,
) (err error) {
	var (
		buf       = make([]byte, startingBufLen)
		bufLen    = len(buf)
		nr        int
		frameSize int
	)

	for {
		// Make sure we have at least a full header
		for nr < stdWriterPrefixLen {
			var nr2 int
			nr2, err = src.Read(buf[nr:])
			nr += nr2
			if errors.Is(err, io.EOF) {
				if nr < stdWriterPrefixLen {
					return nil
				}
				break
			}
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

		var logType tasklog.LogType

		// Check the first byte to know where to write
		stream := buf[stdWriterFdIndex]
		switch stream {
		case Stdin:
			logType = tasklog.LogTypeIn
		case Stdout:
			logType = tasklog.LogTypeOut
		case Stderr, Systemerr:
			logType = tasklog.LogTypeErr
		default:
			return fmt.Errorf("%w: Unrecognized input header: %d", apperrors.ErrInfraInternal, stream)
		}

		// Retrieve the size of the frame
		frameSize = int(binary.BigEndian.Uint32(buf[stdWriterSizeIndex : stdWriterSizeIndex+4]))

		// Check if the buffer is big enough to read the frame.
		// Extend it if necessary.
		if frameSize+stdWriterPrefixLen > bufLen {
			buf = append(buf, make([]byte, frameSize+stdWriterPrefixLen-bufLen+1)...)
			bufLen = len(buf)
		}

		// While the amount of bytes read is less than the size of the frame + header, we keep reading
		for nr < frameSize+stdWriterPrefixLen {
			var nr2 int
			nr2, err = src.Read(buf[nr:])
			nr += nr2
			if errors.Is(err, io.EOF) {
				if nr < frameSize+stdWriterPrefixLen {
					return nil
				}
				break
			}
			if err != nil {
				return apperrors.Wrap(err)
			}
		}

		frameData := buf[stdWriterPrefixLen : frameSize+stdWriterPrefixLen]

		if stream == Stdout && dstOfStdout != nil {
			// Write the retrieved frame (without header)
			nw, err := dstOfStdout.Write(frameData)
			if err != nil {
				return apperrors.Wrap(err)
			}
			// If the frame has not been fully written: error
			if nw != frameSize {
				return apperrors.Wrap(io.ErrShortWrite)
			}
		} else {
			logFrame := &tasklog.LogFrame{
				Type: logType,
				Data: string(frameData),
			}
			if parseTimestamp {
				logFrame.ParseTimestampFromData()
			}
			dst.Send(logFrame)
		}

		// Move the rest of the buffer to the beginning
		copy(buf, buf[frameSize+stdWriterPrefixLen:])
		// Move the index
		nr -= frameSize + stdWriterPrefixLen

		if err = ctx.Err(); err != nil { // Context is done
			return apperrors.Wrap(err)
		}
	}
}
