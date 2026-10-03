package base

type PeriodicKind string

const (
	PeriodicKindHealthCheck PeriodicKind = "healthcheck"
	// PeriodicKindAppAutoscale scales the functions that have autoscale
	// on. HivePaaS makes it and removes it; it is not in AllPeriodicKinds, so a
	// periodic job of this kind cannot be created through the API.
	PeriodicKindAppAutoscale PeriodicKind = "app-autoscale"
)

var (
	AllPeriodicKinds = []PeriodicKind{PeriodicKindHealthCheck}
)
