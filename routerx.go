package routerx

import "net/http"

type Middleware func(http.Handler) http.Handler

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func New() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: nil,
	}
}

func (router *Router) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	router.mux.ServeHTTP(responseWriter, request)
}
