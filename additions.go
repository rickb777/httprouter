package httprouter

import (
	"net/http"
	"strings"
)

// AllMethods is a list of all the 'normal' HTTP methods,
// i.e. HEAD, GET, PUT, POST, DELETE, PATCH, OPTIONS.
//
// It does not include CONNECT or TRACE by default.
// It doesn't include methods used by extension protocols such as WebDav.
// However, you can change it if you need a different set of methods.
var AllMethods = []string{
	http.MethodHead,
	http.MethodGet,
	http.MethodPut,
	http.MethodPost,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodOptions,
}

// HandleAll registers a new request handle with the given path and all methods listed.
// If no methods are specified, then by default AllMethods will be used.
func (r *Router) HandleAll(path string, handle Handle, methods ...string) {
	if len(methods) == 0 {
		methods = AllMethods
	}
	for _, m := range methods {
		r.Handle(m, path, handle)
	}
}

// HandlerAll is an adapter which allows the use of an http.Handler as a
// request handle with a set of methods.
// If no methods are specified, then by default AllMethods will be used.
func (r *Router) HandlerAll(path string, handler http.Handler, methods ...string) {
	if len(methods) == 0 {
		methods = AllMethods
	}
	for _, m := range methods {
		r.Handler(m, path, handler)
	}
}

// SubRouter registers a new request handle with the given path and method(s), trimming
// the prefix from the path before each request is passed to the handle.
//
// The path must end with "/*filepath" (or simply "/*" is allowed in this case). The
// attached handle sees the sub-path only. For example if path is "/a/b/" and the
// request URI path is "/a/b/foo", the handler will see a request for "/foo".
//
// If you don't want the prefix trimmed, instead use Handle with a path that ends with
// ".../*name" (for some name of your choice).
//
// If no methods are specified, all methods (in AllMethods) will be supported. Otherwise,
// only the specified methods will be supported.
func (r *Router) SubRouter(path string, handler http.Handler, methods ...string) {
	if strings.HasSuffix(path, "/*") {
		path = path + "filepath"
	} else if !strings.HasSuffix(path, "/*filepath") {
		panic("'" + path + "' - path must end with /* or /*filepath")
	}

	if strings.IndexByte(path[:len(path)-9], '*') > 0 {
		panic("'" + path + "' - path must contain only one *")
	}

	if len(methods) == 0 {
		methods = AllMethods
	}

	r.HandleAll(path, func(w http.ResponseWriter, req *http.Request, ps Params) {
		req.URL.Path = ps.ByName("filepath")
		handler.ServeHTTP(w, storeParams(ps, req))
	}, methods...)
}

func storeParams(p Params, req *http.Request) *http.Request {
	if len(p) > 0 {
		req = req.WithContext(WithParams(req.Context(), p))
	}
	return req
}

func adapter(handler http.Handler) func(w http.ResponseWriter, req *http.Request, p Params) {
	return func(w http.ResponseWriter, req *http.Request, p Params) {
		handler.ServeHTTP(w, storeParams(p, req))
	}
}
