package router

import (
	"context"
	"net/http"
	"strings"
)

type paramContextKey string
type MiddlewareFunc func(http.Handler) http.Handler

type Router struct {
	routes      []*route
	NotFound    http.Handler
	middlewares []middleware
}

func NewRouter() *Router {
	return &Router{NotFound: http.NotFoundHandler()}
}

func (r *Router) pathSegments(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func (r *Router) Handle(method, pattern string, handler http.Handler) {
	route := &route{
		method:      strings.ToLower(method),
		handler:     handler,
		segs:        r.pathSegments(pattern),
		prefix:      strings.HasSuffix(pattern, "/") || strings.HasSuffix(pattern, "..."),
		queryParams: make(map[string]string),
	}
	r.routes = append(r.routes, route)
}

type middleware interface {
	Middleware(handler http.Handler) http.Handler
}

func (mw MiddlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}

func (r *Router) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		r.middlewares = append(r.middlewares, fn)
	}
}

func (r *Router) UseWithPrefix(prefix string, mwf ...MiddlewareFunc) {
	// Create a new middleware that checks the URL path prefix
	prefixMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Check if the request path has the specified prefix
			if strings.Contains(req.URL.Path, prefix) {
				// Apply the provided middlewares only if the prefix matches
				for _, fn := range mwf {
					next = fn.Middleware(next)
				}
			}
			next.ServeHTTP(w, req)
		})
	}

	// Add the prefix-aware middleware to the global middlewares
	r.middlewares = append(r.middlewares, MiddlewareFunc(prefixMiddleware))
}

func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

func (r *Router) GET(pattern string, fn http.HandlerFunc) {
	r.HandleFunc(http.MethodGet, pattern, fn)
}

func (r *Router) POST(pattern string, fn http.HandlerFunc) {
	r.HandleFunc(http.MethodPost, pattern, fn)
}

func (r *Router) DELETE(pattern string, fn http.HandlerFunc) {
	r.HandleFunc(http.MethodDelete, pattern, fn)
}

func (r *Router) PUT(pattern string, fn http.HandlerFunc) {
	r.HandleFunc(http.MethodPut, pattern, fn)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := strings.ToLower(req.Method)
	segs := r.pathSegments(req.RequestURI)
	notFoundWithMiddleware := r.NotFound
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		notFoundWithMiddleware = r.middlewares[i].Middleware(notFoundWithMiddleware)
	}
	for _, route := range r.routes {
		if route.method != method && route.method != "*" {
			continue
		}
		if ctx, ok := route.match(req.Context(), segs, req); ok {
			handlerWithMiddleware := route.handler
			for i := len(r.middlewares) - 1; i >= 0; i-- {
				handlerWithMiddleware = r.middlewares[i].Middleware(handlerWithMiddleware)
			}
			handlerWithMiddleware.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}
	notFoundWithMiddleware.ServeHTTP(w, req)
}

func Param(ctx context.Context, param string) string {
	// checks if its string
	vStr, ok := ctx.Value(paramContextKey(param)).(string)
	if !ok {
		return ""
	}

	return vStr
}

type route struct {
	method      string
	handler     http.Handler
	prefix      bool
	segs        []string
	queryParams map[string]string
}

func (r *route) match(ctx context.Context, segs []string, req *http.Request) (context.Context, bool) {
	query := req.URL.Query()
	for key, value := range r.queryParams {
		if query.Get(key) != value {
			return nil, false
		}
	}

	if len(segs) > len(r.segs) && !r.prefix {
		return nil, false
	}
	for i, seg := range r.segs {
		if i > len(segs)-1 {
			return nil, false
		}
		isParam := false
		if strings.HasPrefix(seg, ":") {
			isParam = true
			seg = strings.TrimPrefix(seg, ":")
		}
		if !isParam { // verbatim check
			if strings.HasSuffix(seg, "...") {
				if strings.HasPrefix(segs[i], seg[:len(seg)-3]) {
					return ctx, true
				}
			}
			if seg != segs[i] {
				return nil, false
			}
			//if !strings.HasPrefix(segs[i], seg) {
			//	return nil, false
			//}
		}
		if isParam {
			ctx = context.WithValue(ctx, paramContextKey(seg), segs[i])
		}
	}

	return ctx, true
}
