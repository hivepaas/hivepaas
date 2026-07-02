package reflectutil

import "unsafe"

// UnsafeBytesToStr gets string from bytes without copying.
// Use this function for performance purpose, do not modify the byte slice for any reason.
func UnsafeBytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// UnsafeStrToBytes gets the underlying bytes of a string.
// Use this function for performance purpose, do not modify the byte slice for any reason.
func UnsafeStrToBytes(s string) []byte {
	d := unsafe.StringData(s)
	if d == nil {
		return []byte{}
	}
	return unsafe.Slice(d, len(s))
}
