package notificationserviceimpl

import (
	"bytes"
	"sync"
)

const (
	buffSize    = 5000
	buffSizeMax = 64 * 1024 // 64KB max capacity to retain in pool
)

var (
	slackBufPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, buffSize))
		},
	}

	discordBufPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, buffSize))
		},
	}
)

func (s *service) getEmailBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	return s.emailService.GetBuildBuf()
}

func (s *service) getSlackBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = slackBufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			slackBufPool.Put(buf)
		}
	}
}

func (s *service) getDiscordBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = discordBufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			discordBufPool.Put(buf)
		}
	}
}
