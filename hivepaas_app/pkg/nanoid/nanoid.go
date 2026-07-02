package nanoid

import (
	"github.com/jaevor/go-nanoid"
)

var (
	generatorStandard16 = createStandard(16) //nolint:mnd
)

func createStandard(len int) func() string {
	g, err := nanoid.Standard(len)
	if err != nil {
		panic(err)
	}
	return g
}

func NewStandard16() string {
	return generatorStandard16()
}
