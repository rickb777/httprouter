// Copyright 2018 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"net/http"
	"strings"
	"net/url"
)

// StripLeadingSegments drops a fixed number of segments off the front of the URL
// of each request, if it can. For example, if unwantedSegments is 2, a request
// to /a/b/c/d/e will be passed to the handler as /c/d/e.
//
// This is often useful for wrapping handlers added to a Router if they make
// subsequent routing decisions (i.e. some form of nested routing).
//
// See also http.StripPrefix, which strips a fixed prefix instead.
//
// If unwantedSegments is zero, the handler is returned so there is no effect.
func StripLeadingSegments(unwantedSegments uint, handler http.Handler) http.Handler {
	if unwantedSegments == 0 {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, doStripLeadingSegments(r, unwantedSegments))
	})
}

func doStripLeadingSegments(r *http.Request, unwantedSegments uint) *http.Request {
	// make copies so that the originals are unaltered
	r2 := new(http.Request)
	*r2 = *r

	r2.URL = new(url.URL)
	*r2.URL = *r.URL

	p := r.URL.Path
	for unwantedSegments > 0 && len(p) > 0 {
		// the path always starts with leading '/'
		// when received as a request URI
		slash := strings.IndexByte(p[1:], '/')
		if slash > 0 {
			p = p[slash+1:]
		} else {
			p = ""
		}
		unwantedSegments--
	}

	r2.URL.Path = p
	r2.URL.RawPath = "" // throw away original (which is usually blank anyway)
	return r2
}
