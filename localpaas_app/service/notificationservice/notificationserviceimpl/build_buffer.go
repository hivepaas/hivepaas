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
	emailBufPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, buffSize))
		},
	}

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

func getEmailBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = emailBufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			emailBufPool.Put(buf)
		}
	}
}

func getSlackBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = slackBufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			slackBufPool.Put(buf)
		}
	}
}

func getDiscordBuildBuf() (buf *bytes.Buffer, cleanup func()) {
	buf = discordBufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()
	return buf, func() {
		if buf.Cap() <= buffSizeMax {
			discordBufPool.Put(buf)
		}
	}
}
