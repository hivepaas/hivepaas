package base

type CleanupFlag int8

const (
	CleanupFlagFalse = iota
	CleanupFlagTrue
	CleanupFlagForce
)
