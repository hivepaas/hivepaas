package gosample

import (
	"errors"
)

func NewErr(base error, s string) error { return errors.Join(base, errors.New(s)) }

var (
	ErrBase    = errors.New("ERR_BASE")
	ErrDerived = NewErr(ErrBase, "ERR_DERIVED")
)

func use() string {
	return setCustomKey("ERR_REFERENCED_ONLY")
}

func setCustomKey(s string) string { return s }

// notACode is here to prove the scanner is not matching every string.
const notACode = "hello world"

// ErrDeclaredNeverUsed is bound to a name nothing references, the shape of a
// code that was added and then left behind.
var ErrDeclaredNeverUsed = NewErr(ErrBase, "ERR_DECLARED_NEVER_USED")

func alsoUse() error { return ErrDerived }
