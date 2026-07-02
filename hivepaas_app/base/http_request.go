package base

type HTTPMethod string

// Common HTTP methods only
const (
	HTTPMethodGet    HTTPMethod = "GET"
	HTTPMethodPost   HTTPMethod = "POST"
	HTTPMethodPut    HTTPMethod = "PUT"
	HTTPMethodPatch  HTTPMethod = "PATCH"
	HTTPMethodDelete HTTPMethod = "DELETE"
)

var (
	AllHTTPMethods = []HTTPMethod{HTTPMethodGet, HTTPMethodPost, HTTPMethodPut, HTTPMethodPatch,
		HTTPMethodDelete}
)
