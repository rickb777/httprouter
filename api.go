// Copyright 2013 Julien Schmidt & Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.package httprouter

package httprouter

import (
	"net/http"
	"sync"
)

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (path variables).
type Handle func(http.ResponseWriter, *http.Request, Params)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	paramsPool sync.Pool
	maxParams  uint16

	// If enabled, adds the matched route path onto the http.Request context
	// before invoking the handler.
	// The matched route path is only added to handlers of routes that were
	// registered when this option was enabled.
	SaveMatchedRoutePath bool

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 308 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

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

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS http.Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

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

	// PanicHandler is a function to handle panics recovered from http handlers. It should
	// be used to generate an error page and return the http error code 500 (Internal
	// Server Error).
	//
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics. If a panic occurs and this handler is defined, the
	// built-in recover() function obtains the cause and it is passed to the
	// third parameter of this function.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
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

func (r *Router) getParams() *Params {
	ps, _ := r.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (r *Router) putParams(ps *Params) {
	if ps != nil {
		r.paramsPool.Put(ps)
	}
}

func (r *Router) saveMatchedRoutePath(path string, handle Handle) Handle {
	return func(w http.ResponseWriter, req *http.Request, ps Params) {
		if ps == nil {
			psp := r.getParams()
			ps = (*psp)[0:1]
			ps[0] = Param{Key: MatchedRoutePathParam, Value: path}
			handle(w, req, ps)
			r.putParams(psp)
		} else {
			ps = append(ps, Param{Key: MatchedRoutePathParam, Value: path})
			handle(w, req, ps)
		}
	}
}

// GET is a shortcut for router.Handle(http.MethodGet, path, handle)
func (r *Router) GET(path string, handle Handle) {
	r.Handle(http.MethodGet, path, handle)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, path, handle). Note
// that, to follow IETF recommendations, every GET route also provides the
// default handler for equivalent HEAD requests. So it is only necessary to
// register HEAD handlers if they are different from or in addition to the
// GET handlers.
func (r *Router) HEAD(path string, handle Handle) {
	r.Handle(http.MethodHead, path, handle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handle)
func (r *Router) OPTIONS(path string, handle Handle) {
	r.Handle(http.MethodOptions, path, handle)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handle)
func (r *Router) POST(path string, handle Handle) {
	r.Handle(http.MethodPost, path, handle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handle)
func (r *Router) PUT(path string, handle Handle) {
	r.Handle(http.MethodPut, path, handle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handle)
func (r *Router) PATCH(path string, handle Handle) {
	r.Handle(http.MethodPatch, path, handle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handle)
func (r *Router) DELETE(path string, handle Handle) {
	r.Handle(http.MethodDelete, path, handle)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handle Handle) {
	varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handle == nil {
		panic("handle must not be nil")
	}

	if r.SaveMatchedRoutePath {
		varsCount++
		handle = r.saveMatchedRoutePath(path, handle)
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root

		r.globalAllowed = r.allowed("*", "")
	}

	root.addRoute(path, handle)

	// Update maxParams
	if paramsCount := countParams(path); paramsCount+varsCount > r.maxParams {
		r.maxParams = paramsCount + varsCount
	}

	// Lazy-init paramsPool alloc func
	if r.paramsPool.New == nil && r.maxParams > 0 {
		r.paramsPool.New = func() interface{} {
			ps := make(Params, 0, r.maxParams)
			return &ps
		}
	}
}

// Handler is an adapter which allows the usage of an http.Handler as a
// request handle.
// The Params are available in the request context under ParamsKey.
func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path, adapter(handler))
}

// HandlerFunc is an adapter which allows the use of an http.HandlerFunc as a
// request handle.
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handler(method, path, handler)
}

// ServeFiles serves files from the given file system root using the http.FileServer
// handler. Note that http.NotFound is used instead of the Router's NotFound handler;
// if this is inconvenient, consider using SubRouter with your own file server instead
// (see https://github.com/rickb777/servefiles/v3 for example).
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
//
//	router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
//
// Tip: alternative file servers can be plugged in via SubRouter instead.
// This allows, for example, use of the asset handler
// github.com/rickb777/servefiles/v3 with its improved HTTP header
// configuration.
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	r.GET(path, func(w http.ResponseWriter, req *http.Request, ps Params) {
		req.URL.Path = ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})
	// Note that HEAD requests are handled automatically
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
//
// If the path was found, it returns the handler function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (Handle, Params, bool) {
	if root := r.trees[method]; root != nil {
		handle, ps, tsr := root.getValue(path, r.getParams)
		if handle == nil {
			r.putParams(ps)
			return nil, nil, tsr
		}
		if ps == nil {
			return handle, nil, tsr
		}
		return handle, *ps, tsr
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
