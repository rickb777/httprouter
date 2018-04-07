// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"net/http"
)

// Handle registers a new request handle with the given path and method(s).
// If no methods are specified, the default list of methods is "HEAD", "GET".
//
// Usually the respective shortcut functions (GET, POST, PUT etc) can be used
// instead of this method.
func (r *Router) Handle(path string, handle http.Handler, methods ...string) {
	if len(path) == 0 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	if len(methods) == 0 {
		methods = []string{HEAD, GET}
	}

	for _, method := range methods {
		root := r.trees[method]
		if root == nil {
			root = new(node)
			r.trees[method] = root
		}

		root.addRoute(path, handle)
	}
}

// HandleFunc is an adapter which allows the use of an http.HandleFunc as a
// request handle.
// If no methods are specified, the default list of methods is "HEAD", "GET".
func (r *Router) HandleFunc(path string, handler http.HandlerFunc, methods ...string) {
	r.Handle(path, handler, methods...)
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
//
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
//
// The specified handler is used as the file server. This may be nil, in which case
// a http.FileServer is used, but in this case http.NotFound is used instead of
// the Router's NotFound handler.
//
// Both GET and HEAD methods are supported.
//
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"), nil)
func (r *Router) ServeFiles(path string, root http.FileSystem, fileServer http.Handler) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	if fileServer == nil {
		fileServer = http.FileServer(root)
	}

	r.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ps := GetParams(req.Context())
		req.URL.Path = ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	}))
}

func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
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

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}

	path := req.URL.Path

	if root := r.trees[req.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			handle.ServeHTTP(w, req.WithContext(WithParams(req.Context(), ps)))
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed.ServeHTTP(w, req)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}
