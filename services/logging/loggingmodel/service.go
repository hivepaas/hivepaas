package loggingmodel

import "context"

// Backend is the read path. It is required whoever owns the backend, because
// reading is the one thing HivePaaS always does itself.
type Backend interface {
	Query(ctx context.Context, req *QueryReq) (*QueryResp, error)
	// InvocationStats counts a function's invocation lines.
	InvocationStats(ctx context.Context, req *InvocationStatsReq) (*InvocationStatsResp, error)
	// HTTPStats counts an app's requests in the proxy's access log.
	HTTPStats(ctx context.Context, req *HTTPStatsReq) (*HTTPStatsResp, error)
	// ResourceStats reads an app's containers' usage from the agent's rows.
	ResourceStats(ctx context.Context, req *ResourceStatsReq) (*ResourceStatsResp, error)
	Ping(ctx context.Context) error
}

// Collector turns a collection job into the configuration that performs it.
type Collector interface {
	Configure(spec *CollectSpec) (*RuntimeSpec, error)
}

// Deployer describes how to run something.
//
// It is separate from Backend because a backend the user runs needs only
// Backend: HivePaaS queries it and never deploys it. Only a managed component
// implements this.
type Deployer interface {
	RuntimeSpec() (*RuntimeSpec, error)
}
