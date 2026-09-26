package mcp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/mcpuc"
)

// serverInstructions is what a client shows its model about this server as a
// whole, before any tool.
const serverInstructions = "HivePaaS runs apps on a Docker Swarm cluster, grouped in projects, " +
	"each project in envs such as dev or prod. These tools read what the API key's user can see in " +
	"the dashboard: projects, apps, their status and logs, tasks, nodes, the template store and " +
	"scheduled jobs. Name a project, env or app by its key or its name. Nothing here changes anything."

// switchChecker is the part of mcpuc the endpoint needs.
type switchChecker interface {
	IsEnabled(ctx context.Context) (bool, error)
}

// apiKeyAuthenticator is the part of authhandler the endpoint needs.
type apiKeyAuthenticator interface {
	GetAPIKeyAuth(ctx *gin.Context) (*basedto.Auth, error)
	RequestCtx(ctx *gin.Context) context.Context
}

// Services are what the endpoint takes from the rest of the backend. The
// dispatcher, which needs the router, is added by the server.
type Services struct {
	Auth   apiKeyAuthenticator
	Switch switchChecker
	Audit  auditservice.Service
	DB     database.IDB
}

func NewServices(
	auth *authhandler.Handler,
	mcpUC *mcpuc.UC,
	audit auditservice.Service,
	db *database.DB,
) *Services {
	return &Services{Auth: auth, Switch: mcpUC, Audit: audit, DB: db}
}

// Endpoint serves MCP: stateless Streamable HTTP, one server for every request,
// the caller in each request's context.
type Endpoint struct {
	services *Services
	handler  http.Handler
}

func NewEndpoint(services *Services, dispatcher *Dispatcher, tools []Tool) *Endpoint {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "hivepaas", Title: "HivePaaS", Version: base.CurrentVersion},
		&mcpsdk.ServerOptions{Instructions: serverInstructions})
	deps := &Deps{Dispatcher: dispatcher, Audit: services.Audit, DB: services.DB}
	for _, tool := range tools {
		tool.add(server, deps)
	}
	addResources(server, deps)
	addPrompts(server)
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return server },
		&mcpsdk.StreamableHTTPOptions{Stateless: true})
	return &Endpoint{services: services, handler: handler}
}

// Serve is the gin handler of <API base path>/mcp. Off, it is a path like any
// other that does not exist; on, it takes an API key and nothing else.
func (e *Endpoint) Serve(ctx *gin.Context) {
	enabled, err := e.services.Switch.IsEnabled(ctx.Request.Context())
	if err != nil || !enabled {
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	auth, err := e.services.Auth.GetAPIKeyAuth(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "the MCP server takes an API key: the HIVEPAAS-API-KEY-ID and HIVEPAAS-API-SECRET-KEY " +
				"headers, or Authorization: Bearer <keyId>:<secret>. Create one in the dashboard, " +
				"under System settings, AI.",
		})
		return
	}
	reqCtx := withCaller(e.services.Auth.RequestCtx(ctx), auth, ctx.Request)
	e.handler.ServeHTTP(ctx.Writer, ctx.Request.WithContext(reqCtx))
}
