package base

type PeriodicKind string

const (
	PeriodicKindHealthCheck PeriodicKind = "healthcheck"
	PeriodicKindPlaceholder PeriodicKind = "placeholder"
)

var (
	AllPeriodicKinds = []PeriodicKind{PeriodicKindHealthCheck}
)
