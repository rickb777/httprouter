// Copyright 2013 Julien Schmidt & Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.package httprouter

package httprouter

import "net/http"

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
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
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

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
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

// Specify the main HTTP verbs.
const (
	GET     = "GET"
	PUT     = "PUT"
	HEAD    = "HEAD"
	POST    = "POST"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

// GET is a shortcut for router.Handle(path, handle, "GET")
func (r *Router) GET(path string, handle http.Handler) {
	r.Handle(path, handle, GET)
}

// HEAD is a shortcut for router.Handle(path, handle, "HEAD")
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

// AllMethods is a list of all the 'normal' methods,
// i.e. HEAD, GET, PUT, POST, DELETE, PATCH, OPTIONS.
//
// It doesn't include methods used by extension protocols such as WebDav, although
// you can change it if you need this.
var AllMethods = []string{HEAD, GET, PUT, POST, DELETE, PATCH, OPTIONS}

// HandleAll registers a new request handle with the given path and all method listed
// in AllMethods.
func (r *Router) HandleAll(path string, handle http.Handler) {
	r.Handle(path, handle, AllMethods...)
}
