// Copyright 2013 Julien Schmidt & Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.package httprouter

package httprouter

import (
	"net/http"
	"strings"
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handler is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	//
	// If a handler can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// SpecialisedHEAD allows HEAD routes to be different from the GET routes.
	// The default behaviour, when this flag is false, is for all HEAD requests
	// to automatically use the same routing rules defined for GET.
	SpecialisedHEAD bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// Configurable http.Handler which is called when no matching route is
	// found. Also use this if you need to cascade to another router (perhaps
	// via intermediate middleware). If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// A function to handle panics recovered from http handlers. It should be used
	// to generate an error page and return the http error code 500 (Internal
	// Server Error).
	//
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics. If a panic occurs and this handler is defined, the
	// built-in recover() function obtains the cause and it is passed to the
	// third parameter of this function.
	PanicHandler func(w http.ResponseWriter, req *http.Request, rcv interface{})
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New()

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

// Specify the main HTTP verbs.
const (
	GET     = "GET"
	PUT     = "PUT"
	HEAD    = "HEAD"
	POST    = "POST"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	CONNECT = "CONNECT"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE" // a diganostic method rarely used
)

// GET is a shortcut for router.Handle(path, handle, "GET")
func (r *Router) GET(path string, handle http.Handler) {
	r.Handle(path, handle, GET)
}

// HEAD is a shortcut for router.Handle(path, handle, "HEAD")
// Note that Router.SpecialisedHEAD flag must be set true;
// otherwise the route defined here will be ignored.
func (r *Router) HEAD(path string, handle http.Handler) {
	r.Handle(path, handle, HEAD)
}

// OPTIONS is a shortcut for router.Handle(path, handle, "OPTIONS")
func (r *Router) OPTIONS(path string, handle http.Handler) {
	r.Handle(path, handle, OPTIONS)
}

// POST is a shortcut for router.Handle(path, handle, "POST")
func (r *Router) POST(path string, handle http.Handler) {
	r.Handle(path, handle, POST)
}

// PUT is a shortcut for router.Handle(path, handle, "PUT")
func (r *Router) PUT(path string, handle http.Handler) {
	r.Handle(path, handle, PUT)
}

// PATCH is a shortcut for router.Handle(path, handle, "PATCH")
func (r *Router) PATCH(path string, handle http.Handler) {
	r.Handle(path, handle, PATCH)
}

// DELETE is a shortcut for router.Handle(path, handle, "DELETE")
func (r *Router) DELETE(path string, handle http.Handler) {
	r.Handle(path, handle, DELETE)
}

// AllMethods is a list of all the 'normal' HTTP methods,
// i.e. HEAD, GET, PUT, POST, DELETE, PATCH, OPTIONS.
//
// It does not include CONNECT or TRACE by default.
// It doesn't include methods used by extension protocols such as WebDav.
// However, you can change it if you need a different set of methods.
var AllMethods = []string{HEAD, GET, PUT, POST, DELETE, PATCH, OPTIONS}

// HandleAll registers a new request handle with the given path and all method listed
// in AllMethods.
func (r *Router) HandleAll(path string, handler http.Handler) {
	r.Handle(path, handler, AllMethods...)
}

// Handle registers a new request handler with the given path and method(s).
// If no methods are specified, "GET" only is assumed (although HEAD is also
// implicitly supported unless Router.SpecialisedHEAD is set).
//
// Usually the respective shortcut functions (GET, POST, PUT etc) can be used
// instead of this method.
//
// The handler sees the original request URI unaltered; see also SubRouter for
// a different capability.
func (r *Router) Handle(path string, handler http.Handler, methods ...string) {
	if len(path) == 0 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	if len(methods) == 0 {
		methods = []string{GET}
	}

	for _, method := range methods {
		root := r.trees[method]
		if root == nil {
			root = new(node)
			r.trees[method] = root
		}

		root.addRoute(path, handler)
	}
}

// HandleFunc is an adapter which allows the use of an http.HandleFunc as a
// request handler.
//
// If no methods are specified, "GET" only is assumed (although HEAD is also
// implicitly supported unless Router.SpecialisedHEAD is set).
func (r *Router) HandleFunc(path string, handler http.HandlerFunc, methods ...string) {
	r.Handle(path, handler, methods...)
}

// SubRouter registers a new request handler with the given path and method(s), trimming
// the prefix from the path before each request is passed to the handler.
//
// The path must end with "/*filepath" (or simply "/*" is allowed in this case). The
// attached handler sees the sub-path only. For example if path is "/a/b/" and the
// request URI path is "/a/b/foo", the handler will see a request for "/foo".
//
// If no methods are specified, all methods will be supported. Otherwise, only the
// specified methods will be supported.
//
// If you don't want the prefix trimmed, instead use Handle with a path that ends with
// ".../*name" (for some name of your choice).
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

	r.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ps := GetParams(req.Context())
		req.URL.Path = ps.ByName("filepath")
		handler.ServeHTTP(w, req)
	}), methods...)
}

// ServeFiles serves files from the given file system root using the http.FileServer
// handler. Note that http.NotFound is used instead of the Router's NotFound handler;
// if this is inconvenient, consider using SubRouter with your own file server instead.
//
// The path must end with "/*filepath" (or simply "/*" is allowed in this case), files
// are then served from the local path /defined/root/dir/*filepath.
//
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
//
// Both GET and HEAD methods are supported, but no other methods.
//
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	methods := []string{GET}
	if r.SpecialisedHEAD {
		methods = []string{GET, HEAD}
	}
	r.SubRouter(path, http.FileServer(root), methods...)
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
//
// If the path was found, it returns the handler function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (http.Handler, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

// ListPaths allows inspection of the paths known to the router, grouped by method.
// If method is blank, all registered methods are returned.
//
// The resulting slices are sorted in increasing order.
//
// This is intended for debugging and diagnostics.
func (r *Router) ListPaths(method string) map[string][]string {
	result := make(map[string][]string)
	if method == "" {
		for m, root := range r.trees {
			result[m] = root.makePathList(nil, nil)
		}
	} else {
		if root := r.trees[method]; root != nil {
			result[method] = root.makePathList(nil, nil)
		}
	}
	return result
}
