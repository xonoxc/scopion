package app

import (
	"net/http"

	"github.com/xonoxc/scopion/internal/api"
	"github.com/xonoxc/scopion/internal/api/middleware"
	"github.com/xonoxc/scopion/internal/ingest"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/pipeline"
	"github.com/xonoxc/scopion/internal/store"
)

type AppRouter struct {
	store       store.Storage
	broadcaster *live.Broadcaster
	pipeline    *pipeline.Batcher
	config      ServerConfig
}

func NewAppRouter(s store.Storage, broadcaster *live.Broadcaster, bp *pipeline.Batcher, config ServerConfig) *AppRouter {
	return &AppRouter{
		store:       s,
		broadcaster: broadcaster,
		pipeline:    bp,
		config:      config,
	}
}

type Route struct {
	Path    string
	Handler http.Handler
}

func (a *AppRouter) getRoutes() []Route {
	return []Route{
		{Path: "/api/live", Handler: live.SSE(a.broadcaster)},
		{Path: "/api/events", Handler: api.EventsHandler(a.store)},
		{Path: "/api/trace-events", Handler: api.TraceEventsHandler(a.store)},
		{Path: "/api/trace-spans", Handler: api.TraceSpansHandler(a.store)},
		{Path: "/api/stats", Handler: api.StatsHandler(a.store)},
		{Path: "/api/throughput", Handler: api.ThroughputHandler(a.store)},
		{Path: "/api/errors-by-service", Handler: api.ErrorsByServiceHandler(a.store)},
		{Path: "/api/services", Handler: api.ServicesHandler(a.store)},
		{Path: "/api/traces", Handler: api.TracesHandler(a.store)},
		{Path: "/api/search", Handler: api.SearchHandler(a.store)},
		{Path: "/api/error-groups", Handler: api.ErrorGroupsHandler(a.store)},
		{Path: "/api/status", Handler: api.StatusHandler(a.config.IsDemoMode())},
		{Path: "/ingest", Handler: ingest.Handler(a.pipeline)},
		{Path: "/ingest-span", Handler: ingest.SpanHandler(a.pipeline)},
	}
}

func (a *AppRouter) Setup() *http.ServeMux {
	mux := http.NewServeMux()
	routes := a.getRoutes()

	for _, r := range routes {
		mux.Handle(r.Path, middleware.LoggingMiddleware(r.Handler))
	}

	return mux
}
