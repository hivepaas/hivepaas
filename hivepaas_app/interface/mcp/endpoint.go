package mcp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/mcpuc"
)

// readInstructions is what a client shows its model about this server as a
// whole, before any tool, when the caller may only read.
const readInstructions = "HivePaaS runs apps on a Docker Swarm cluster, grouped in projects, " +
	"each project in envs such as dev or prod. These tools read what the API key's user can see in " +
	"the dashboard: projects, apps, their status and logs, tasks, nodes, the template store and " +
	"scheduled jobs. Name a project, env or app by its key or its name. Nothing here changes anything."

// writeInstructions is the same, for a caller who may change things.
const writeInstructions = "HivePaaS runs apps on a Docker Swarm cluster, grouped in projects, " +
	"each project in envs such as dev or prod. These tools read what the API key's user can see in " +
	"the dashboard - projects, apps, their status and logs, tasks, nodes, the template store and " +
	"scheduled jobs - and change some of it: install, restart or redeploy an app, change its " +
	"configuration, schedule a job. Name a project, env or app by its key or its name. " +
	"Tools named plan_* change nothing: each answers a plan. Show the person the plan in full, and " +
	"call apply_plan with its planToken only once they have agreed to it. Never apply a plan the " +
	"person has not seen."

// settingsReader is the part of mcpuc the endpoint needs.
type settingsReader interface {
	Current(ctx context.Context) (entity.MCPSettings, error)
}

// apiKeyAuthenticator is the part of authhandler the endpoint needs.
type apiKeyAuthenticator interface {
	GetAPIKeyAuth(ctx *gin.Context) (*basedto.Auth, string, error)
	RequestCtx(ctx *gin.Context) context.Context
}

// Services are what the endpoint takes from the rest of the backend. The
// dispatcher, which needs the router, is added by the server.
type Services struct {
	Auth     apiKeyAuthenticator
	Settings settingsReader
	Plans    planRepo
	Audit    auditservice.Service
	DB       database.IDB
}

func NewServices(
	auth *authhandler.Handler,
	mcpUC *mcpuc.UC,
	plans cacherepository.MCPPlanRepo,
	audit auditservice.Service,
	db *database.DB,
) *Services {
	return &Services{Auth: auth, Settings: mcpUC, Plans: plans, Audit: audit, DB: db}
}

// Endpoint serves MCP: stateless Streamable HTTP, the caller in each request's
// context. It keeps two servers, and serves each request the one its caller may
// use: a client is never listed a tool it cannot call.
type Endpoint struct {
	services *Services
	handler  http.Handler
}

func NewEndpoint(services *Services, dispatcher *Dispatcher, tools []Tool) *Endpoint {
	deps := &Deps{Dispatcher: dispatcher, Audit: services.Audit, DB: services.DB, Settings: services.Settings,
		Plans: services.Plans, appliers: map[string]*applier{}}
	for _, tool := range tools {
		if tool.applies != nil {
			deps.appliers[tool.Name] = tool.applies
		}
	}
	servers := make(map[access]*mcpsdk.Server, len(allAccesses))
	for _, a := range allAccesses {
		servers[a] = newServer(deps, tools, a)
	}
	handler := mcpsdk.NewStreamableHTTPHandler(func(r *http.Request) *mcpsdk.Server {
		if c := callerFrom(r.Context()); c != nil {
			return servers[c.access]
		}
		return servers[access{}]
	}, &mcpsdk.StreamableHTTPOptions{Stateless: true})
	return &Endpoint{services: services, handler: handler}
}

// newServer is the server of one access: the tools it can use, and what it
// is told of those it cannot.
func newServer(deps *Deps, tools []Tool, a access) *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "hivepaas", Title: "HivePaaS", Version: base.CurrentVersion},
		&mcpsdk.ServerOptions{Instructions: a.instructions()})
	for _, tool := range tools {
		if a.serves(tool.needs) {
			tool.add(server, deps)
		}
	}
	addResources(server, deps)
	addPrompts(server, a)
	return server
}

// Serve is the gin handler of <API base path>/mcp. Off, it is a path like any
// other that does not exist; on, it takes an API key and nothing else.
func (e *Endpoint) Serve(ctx *gin.Context) {
	current, err := e.services.Settings.Current(ctx.Request.Context())
	if err != nil || !current.Enabled {
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	auth, keyID, err := e.services.Auth.GetAPIKeyAuth(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "the MCP server takes an API key: the HIVEPAAS-API-KEY-ID and HIVEPAAS-API-SECRET-KEY " +
				"headers, or Authorization: Bearer <keyId>:<secret>. Create one in the dashboard, " +
				"under Profile, API keys.",
		})
		return
	}
	c := &caller{auth: auth, keyID: keyID, access: accessOf(current.AllowWrite, auth)}
	reqCtx := withCaller(e.services.Auth.RequestCtx(ctx), c, ctx.Request)
	e.handler.ServeHTTP(ctx.Writer, ctx.Request.WithContext(reqCtx))
}
