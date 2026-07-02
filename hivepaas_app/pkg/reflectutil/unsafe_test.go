package reflectutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnsafeBytesToStr(t *testing.T) {
	assert.Equal(t, "", UnsafeBytesToStr(nil))
	assert.Equal(t, "", UnsafeBytesToStr([]byte{}))
	assert.Equal(t, "abc 123", UnsafeBytesToStr([]byte("abc 123")))
}

func Test_UnsafeStrToBytes(t *testing.T) {
	assert.Equal(t, []byte{}, UnsafeStrToBytes(""))
	assert.Equal(t, []byte("abc 123"), UnsafeStrToBytes("abc 123"))
}
