package routerx

import (
	"net/http"
	"strings"
)

// Middleware wraps an http.Handler and returns another http.Handler.
// It is typically used for cross-cutting concerns such as logging, recovery,
// authentication, or metrics.
type Middleware func(http.Handler) http.Handler

// Router is the main entry point to routerx.
// It wraps the standard library http.ServeMux and adds support for:
//
//   - HTTP method-aware patterns (GET /path, POST /path, etc.)
//   - Route groups with path prefixes
//   - Fluent path builders
//   - Middleware chains at router, group, and path levels
//
// Router implements http.Handler and can be passed directly to http.ListenAndServe.
type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

// RouteGroup represents a group of routes that share a common path prefix
// and a shared middleware chain. Nested groups inherit and extend the
// middleware of their parent groups.
type RouteGroup struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []Middleware
}

// PathBuilder provides a fluent API for registering multiple HTTP methods
// for a single path. It inherits middlewares from the router or group that
// created it and applies them to each registered handler.
type PathBuilder struct {
	mux         *http.ServeMux
	basePath    string
	middlewares []Middleware
}

// New creates a new Router using the standard library http.ServeMux as the
// underlying multiplexer. The returned Router is empty and ready for route
// registration.
func New() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: nil,
	}
}

// ServeHTTP makes Router implement http.Handler. Incoming requests are passed
// directly to the underlying http.ServeMux after all routes have been registered.
func (router *Router) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	router.mux.ServeHTTP(responseWriter, request)
}

// Use appends one or more Middleware instances to the Router.
// All routes registered on this router after calling Use will use the
// accumulated middleware chain. Use returns the Router to support chaining.
//
// Example:
//
//	router := routerx.New().
//	    Use(LoggingMiddleware, RecoveryMiddleware)
func (router *Router) Use(middlewares ...Middleware) *Router {
	router.middlewares = append(router.middlewares, middlewares...)
	return router
}

// Group creates a new RouteGroup with the given path prefix. The group
// inherits all middlewares currently configured on the Router. Routes
// registered on the group will be rooted at the prefix.
//
// Example:
//
//	api := router.Group("/api")
//	api.Get("/status", statusHandler) // matches GET /api/status
func (router *Router) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		mux:         router.mux,
		prefix:      cleanPath(prefix),
		middlewares: copyMiddlewares(router.middlewares),
	}
}

// Path creates a PathBuilder for the given path. The builder inherits all
// middlewares currently configured on the Router. This is useful when you
// want to register multiple HTTP methods for the same path.
//
// Example:
//
//	router.Path("/users").
//	    Get(listUsers).
//	    Post(createUser)
func (router *Router) Path(path string) *PathBuilder {
	fullPath := cleanPath(path)
	return &PathBuilder{
		mux:         router.mux,
		basePath:    fullPath,
		middlewares: copyMiddlewares(router.middlewares),
	}
}

func (router *Router) Get(path string, handler http.HandlerFunc) {
	router.handle("GET", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Post(path string, handler http.HandlerFunc) {
	router.handle("POST", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Patch(path string, handler http.HandlerFunc) {
	router.handle("PATCH", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Delete(path string, handler http.HandlerFunc) {
	router.handle("DELETE", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Head(path string, handler http.HandlerFunc) {
	router.handle("HEAD", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Put(path string, handler http.HandlerFunc) {
	router.handle("PUT", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Options(path string, handler http.HandlerFunc) {
	router.handle("OPTIONS", cleanPath(path), handler, router.middlewares)
}

func (router *Router) Connect(path string, handler http.HandlerFunc) {
	router.handle("CONNECT", cleanPath(path), handler, router.middlewares)
}

// Trace registers a handler for HTTP TRACE requests at the specified path.
func (router *Router) Trace(path string, handler http.HandlerFunc) {
	router.handle("TRACE", cleanPath(path), handler, router.middlewares)
}

func (router *Router) handle(method string, path string, handler http.HandlerFunc, middlewares []Middleware) {
	pattern := method + " " + path
	finalHandler := applyMiddlewares(http.HandlerFunc(handler), middlewares)
	router.mux.Handle(pattern, finalHandler)
}

// Use appends one or more Middleware instances to the RouteGroup.
// Middlewares added to the group are applied after the Router middlewares
// but before any PathBuilder-specific middlewares. Use returns the group
// to support chaining.
func (group *RouteGroup) Use(middlewares ...Middleware) *RouteGroup {
	group.middlewares = append(group.middlewares, middlewares...)
	return group
}

// Group creates a nested RouteGroup under the current group's prefix.
// The nested group inherits the group's prefix and middleware chain.
//
// Example:
//
//	api := router.Group("/api")
//	v1 := api.Group("/v1")
//	v1.Get("/users", handler) // matches GET /api/v1/users
func (group *RouteGroup) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		mux:         group.mux,
		prefix:      joinPath(group.prefix, prefix),
		middlewares: copyMiddlewares(group.middlewares),
	}
}

// Path creates a PathBuilder rooted at the group's prefix. The builder
// inherits the group's middleware chain.
//
// Example:
//
//	apiV1 := router.Group("/api").Group("/v1")
//	apiV1.Path("/users").
//	    Get(listUsers).
//	    Post(createUser)
func (group *RouteGroup) Path(path string) *PathBuilder {
	fullPath := joinPath(group.prefix, path)
	return &PathBuilder{
		mux:         group.mux,
		basePath:    fullPath,
		middlewares: copyMiddlewares(group.middlewares),
	}
}

func (group *RouteGroup) Get(path string, handler http.HandlerFunc) {
	group.handle("GET", path, handler)
}

func (group *RouteGroup) Post(path string, handler http.HandlerFunc) {
	group.handle("POST", path, handler)
}

func (group *RouteGroup) Patch(path string, handler http.HandlerFunc) {
	group.handle("PATCH", path, handler)
}

func (group *RouteGroup) Delete(path string, handler http.HandlerFunc) {
	group.handle("DELETE", path, handler)
}
func (group *RouteGroup) Head(path string, handler http.HandlerFunc) {
	group.handle("HEAD", path, handler)
}

func (group *RouteGroup) Put(path string, handler http.HandlerFunc) {
	group.handle("PUT", path, handler)
}

func (group *RouteGroup) Options(path string, handler http.HandlerFunc) {
	group.handle("OPTIONS", path, handler)
}

func (group *RouteGroup) Connect(path string, handler http.HandlerFunc) {
	group.handle("CONNECT", path, handler)
}

func (group *RouteGroup) Trace(path string, handler http.HandlerFunc) {
	group.handle("TRACE", path, handler)
}

func (group *RouteGroup) handle(method string, path string, handler http.HandlerFunc) {
	fullPath := joinPath(group.prefix, path)
	pattern := method + " " + fullPath
	finalHandler := applyMiddlewares(http.HandlerFunc(handler), group.middlewares)
	group.mux.Handle(pattern, finalHandler)
}

func (builder *PathBuilder) Get(handler http.HandlerFunc) *PathBuilder {
	builder.register("GET", handler)
	return builder
}

func (builder *PathBuilder) Post(handler http.HandlerFunc) *PathBuilder {
	builder.register("POST", handler)
	return builder
}

func (builder *PathBuilder) Patch(handler http.HandlerFunc) *PathBuilder {
	builder.register("PATCH", handler)
	return builder
}

func (builder *PathBuilder) Delete(handler http.HandlerFunc) *PathBuilder {
	builder.register("DELETE", handler)
	return builder
}

func (builder *PathBuilder) register(method string, handler http.HandlerFunc) {
	pattern := method + " " + builder.basePath
	finalHandler := applyMiddlewares(http.HandlerFunc(handler), builder.middlewares)
	builder.mux.Handle(pattern, finalHandler)
}
func (builder *PathBuilder) Head(handler http.HandlerFunc) *PathBuilder {
	builder.register("HEAD", handler)
	return builder
}

func (builder *PathBuilder) Put(handler http.HandlerFunc) *PathBuilder {
	builder.register("PUT", handler)
	return builder
}

func (builder *PathBuilder) Options(handler http.HandlerFunc) *PathBuilder {
	builder.register("OPTIONS", handler)
	return builder
}

func (builder *PathBuilder) Connect(handler http.HandlerFunc) *PathBuilder {
	builder.register("CONNECT", handler)
	return builder
}

func (builder *PathBuilder) Trace(handler http.HandlerFunc) *PathBuilder {
	builder.register("TRACE", handler)
	return builder
}

// applyMiddlewares applies a slice of middlewares to the provided handler.
// Middlewares are applied in the order they were added: the first middleware
// in the slice becomes the outermost wrapper.
func applyMiddlewares(handler http.Handler, middlewares []Middleware) http.Handler {
	if len(middlewares) == 0 {
		return handler
	}
	for index := len(middlewares) - 1; index >= 0; index-- {
		handler = middlewares[index](handler)
	}
	return handler
}

// copyMiddlewares returns a shallow copy of the provided middleware slice.
// It is used to ensure that groups and path builders do not accidentally
// share the same backing array and can safely append their own middlewares.
func copyMiddlewares(middlewares []Middleware) []Middleware {
	if len(middlewares) == 0 {
		return nil
	}
	result := make([]Middleware, len(middlewares))
	copy(result, middlewares)
	return result
}

// cleanPath normalizes a route path by ensuring that it starts with a slash
// and does not end with a trailing slash (unless it is the root path).
func cleanPath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(path, "/")
}

// joinPath concatenates a prefix and a path into a single clean path,
// taking care of leading and trailing slashes.
func joinPath(prefix string, path string) string {
	if prefix == "" || prefix == "/" {
		return cleanPath(path)
	}
	return cleanPath(prefix) + cleanPath(path)
}
