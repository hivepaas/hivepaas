package docker

import (
	"context"
	"encoding/json"
	"io"

	"github.com/moby/moby/api/types/jsonstream"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

type JSONMsg struct {
	*jsonstream.Message
}

func (msg *JSONMsg) String() string {
	if msg.Stream != "" {
		return msg.Stream
	}
	return msg.ErrorStr()
}

func (msg *JSONMsg) ErrorStr() string {
	if msg.Error != nil {
		return msg.Error.Error()
	}
	return ""
}

func StartScanningJSONMsg(
	ctx context.Context,
	reader io.ReadCloser,
	options batchrecvchan.Options, // if zero, scan one by one
) (msgChan <-chan []*JSONMsg, closeFunc func() error) {
	batchChan := batchrecvchan.NewChan[*JSONMsg](options)

	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		context.AfterFunc(ctx, func() { _ = reader.Close() })
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				safego.LogPanic("docker.scanJSONMsg", r)
			}
			batchChan.Close()
		}()

		// Close logs stream
		defer reader.Close()

		decoder := json.NewDecoder(reader)
		for {
			var jm jsonstream.Message
			err := decoder.Decode(&jm)
			if err != nil {
				// Any decode error is terminal, not just io.EOF: retrying on a
				// broken stream would spin forever emitting empty messages.
				break
			}
			msg := &JSONMsg{Message: &jm}

			if ctx.Err() != nil { // context is done
				return
			}

			batchChan.Send(msg)
		}
	}()

	return batchChan.Receiver(), func() error { return reader.Close() }
}
