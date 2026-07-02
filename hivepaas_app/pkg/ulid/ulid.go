package ulid

import (
	"crypto/rand"
	"sync"
	"time"

	ulidLib "github.com/oklog/ulid/v2"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

var (
	entropy    = ulidLib.Monotonic(rand.Reader, 0)
	entropyMtx sync.Mutex
	nowFunc    = time.Now
)

// NewULID creates a new 16-byte ULID object
func NewULID() (ulidLib.ULID, error) {
	entropyMtx.Lock()
	defer entropyMtx.Unlock()
	newULID, err := ulidLib.New(ulidLib.Timestamp(nowFunc()), entropy)
	if err != nil {
		return ulidLib.ULID{}, tracerr.Wrap(err)
	}
	return newULID, nil
}

// NewStringULID creates a new ULID as 26-char string
func NewStringULID() (string, error) {
	ulid, err := NewULID()
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	var newID ulidLib.ULID
	err = newID.UnmarshalBinary(ulid[:])
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	return newID.String(), nil
}

func ParseULID(id string) (ulidLib.ULID, error) {
	return ulidLib.Parse(id) //nolint:wrapcheck
}

func IsULID(id string) bool {
	_, err := ParseULID(id)
	return err == nil
}
