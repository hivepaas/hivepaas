package bbpool

import "github.com/valyala/bytebufferpool"

var (
	mediumPool bytebufferpool.Pool
	largePool  bytebufferpool.Pool
)

// Small - For buffers less than 64KB (uses the library's default pool)
func Small() (*bytebufferpool.ByteBuffer, func(*bytebufferpool.ByteBuffer)) {
	return bytebufferpool.Get(), bytebufferpool.Put
}

// Medium - For buffers less than 1MB
func Medium() (*bytebufferpool.ByteBuffer, func(*bytebufferpool.ByteBuffer)) {
	return mediumPool.Get(), PutMedium
}

func PutMedium(b *bytebufferpool.ByteBuffer) {
	mediumPool.Put(b)
}

// Large - For buffers less than 100MB
func Large() (*bytebufferpool.ByteBuffer, func(*bytebufferpool.ByteBuffer)) {
	return largePool.Get(), PutLarge
}

func PutLarge(b *bytebufferpool.ByteBuffer) {
	largePool.Put(b)
}
