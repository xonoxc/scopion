package app

import (
	"net/http"

	"github.com/xonoxc/scopion/internal/api"
	"github.com/xonoxc/scopion/internal/api/middleware"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/ingest"
	"github.com/xonoxc/scopion/internal/live"
)

type AppRouter struct {
	appState    *appcontext.AtomicAppState
	broadcaster *live.Broadcaster
	config      ServerConfig
}

func NewAppRouter(appState *appcontext.AtomicAppState, broadcaster *live.Broadcaster, config ServerConfig) *AppRouter {
	return &AppRouter{
		appState:    appState,
		broadcaster: broadcaster,
		config:      config,
	}
}

type Route struct {
	Path       string
	Handler    http.Handler
	Middleware []func(http.Handler) http.Handler
}

/*
* Global middlewares here
***/
func (a *AppRouter) globalMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.LoggingMiddleware,
	}
}

func (a *AppRouter) getRoutes() []Route {
	return []Route{
		{Path: "/api/live", Handler: live.SSE(a.broadcaster)},
		{Path: "/api/events", Handler: api.EventsHandler(a.appState)},
		{Path: "/api/trace-events", Handler: api.TraceEventsHandler(a.appState)},
		{Path: "/api/stats", Handler: api.StatsHandler(a.appState)},
		{Path: "/api/throughput", Handler: api.ThroughputHandler(a.appState)},
		{Path: "/api/errors-by-service", Handler: api.ErrorsByServiceHandler(a.appState)},
		{Path: "/api/services", Handler: api.ServicesHandler(a.appState)},
		{Path: "/api/traces", Handler: api.TracesHandler(a.appState)},
		{Path: "/api/search", Handler: api.SearchHandler(a.appState)},
		{Path: "/api/status", Handler: api.StatusHandler(a.config.IsDemoMode())},
		{Path: "/ingest", Handler: ingest.Handler(a.appState.Snapshot().Store, a.broadcaster)},
	}
}

func (a *AppRouter) Setup() {
	routes := a.getRoutes()
	globalsMids := a.globalMiddleware()

	for _, r := range routes {
		h := r.Handler

		for i := len(r.Middleware) - 1; i >= 0; i-- {
			h = r.Middleware[i](h)
		}

		for i := len(globalsMids) - 1; i >= 0; i-- {
			h = globalsMids[i](h)
		}

		http.Handle(r.Path, middleware.LoggingMiddleware(h))
	}
}
