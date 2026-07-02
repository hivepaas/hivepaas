package base

import "google.golang.org/grpc/health/grpc_health_v1"

type HealthcheckType string

const (
	HealthcheckTypeREST HealthcheckType = "rest"
	HealthcheckTypeGRPC HealthcheckType = "grpc"
)

var (
	AllHealthcheckTypes = []HealthcheckType{HealthcheckTypeREST, HealthcheckTypeGRPC}
)

type HealthcheckGRPCVersion string

const (
	HealthcheckGRPCV1 HealthcheckGRPCVersion = "v1"
)

var (
	AllHealthcheckGRPCVersions = []HealthcheckGRPCVersion{HealthcheckGRPCV1}
)

type HealthcheckGRPCStatus int32

const (
	HealthcheckGRPCV1StatusUnknown        = HealthcheckGRPCStatus(grpc_health_v1.HealthCheckResponse_UNKNOWN)
	HealthcheckGRPCV1StatusServing        = HealthcheckGRPCStatus(grpc_health_v1.HealthCheckResponse_SERVING)
	HealthcheckGRPCV1StatusNotServing     = HealthcheckGRPCStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	HealthcheckGRPCV1StatusServiceUnknown = HealthcheckGRPCStatus(grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN)
)

var (
	AllHealthcheckGRPCV1Statuses = []HealthcheckGRPCStatus{HealthcheckGRPCV1StatusUnknown,
		HealthcheckGRPCV1StatusServing, HealthcheckGRPCV1StatusNotServing, HealthcheckGRPCV1StatusServiceUnknown}
)
