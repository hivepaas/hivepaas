// NOTE: source copied from https://github.com/simonfrey/jsonl

package jsonl

import (
	"encoding/json"
	"fmt"
	"io"
)

type Writer struct {
	w io.Writer
}

type WriterOption func(w *Writer)

func NewWriter(w io.Writer, opts ...WriterOption) *Writer {
	wr := &Writer{
		w: w,
	}
	for _, opt := range opts {
		opt(wr)
	}
	return wr
}

func (w Writer) Close() error {
	if c, ok := w.w.(io.WriteCloser); ok {
		return c.Close() //nolint:wrapcheck
	}
	return ErrNotCloseable
}

func (w Writer) Write(data any) error {
	j, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not json marshal data: %w", err)
	}

	_, err = w.w.Write(j)
	if err != nil {
		return fmt.Errorf("could not write json data to underlying io.Writer: %w", err)
	}

	_, err = w.w.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("could not write newline to underlying io.Writer: %w", err)
	}

	return nil
}

func (w Writer) WriteMetadata(metadata any) error {
	return w.Write(metadata)
}

func (w Writer) WriteChunk(chunk any) error {
	return w.Write(chunk)
}
