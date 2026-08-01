package base

type PeriodicKind string

const (
	PeriodicKindHealthCheck PeriodicKind = "healthcheck"
)

var (
	AllPeriodicKinds = []PeriodicKind{PeriodicKindHealthCheck}
)
