package emailserviceimpl

import (
	"bytes"
	"sync"
)

const (
	buffSize    = 5000
	buffSizeMax = 64 * 1024 // 64KB max capacity to retain in pool
)

var (
	bufPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, buffSize))
		},
	}
)

func (s *service) getBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = bufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			bufPool.Put(buf)
		}
	}
}

func (s *service) GetBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	return s.getBuildBuf()
}
