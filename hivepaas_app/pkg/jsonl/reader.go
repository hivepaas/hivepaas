// NOTE: source copied from https://github.com/simonfrey/jsonl

package jsonl

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

var (
	ErrNotCloseable       = errors.New("not closeable")
	ErrScannerNotReadable = errors.New("could not read from scanner")
)

type Reader struct {
	r       io.Reader
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) Reader {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	return Reader{
		r:       r,
		scanner: scanner,
	}
}

func (r Reader) Close() error {
	if c, ok := r.r.(io.ReadCloser); ok {
		return c.Close() //nolint:wrapcheck
	}
	return ErrNotCloseable
}

func (r Reader) ReadSingleLine(output any) error {
	ok := r.scanner.Scan()
	if !ok {
		return ErrScannerNotReadable
	}
	err := json.Unmarshal(r.scanner.Bytes(), output)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (r Reader) ReadLines(callback func(data []byte) error) error {
	for r.scanner.Scan() {
		err := callback(r.scanner.Bytes())
		if err != nil {
			return fmt.Errorf("error in callback: %w", err)
		}
	}
	return nil
}
