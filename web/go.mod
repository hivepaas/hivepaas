// Not Go: a module of its own only so that the go tool (go test ./...,
// golangci-lint, govulncheck) leaves the dashboard out of the backend's module,
// with the Go files some npm packages carry in node_modules.
module github.com/hivepaas/hivepaas/web
