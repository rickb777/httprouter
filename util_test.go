// Copyright 2018 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"net/http/httptest"
	"testing"
	"net/url"
)

func TestStripLeadingSegments_should_create_handler_that_drops_segments(t *testing.T) {

	var cases = []struct {
		url, remaining string
		strip          uint
		mustDiffer     bool
	}{
		{"/a/123/z", "/a/123/z", 0, false},
		{"/a/123/z", "/123/z", 1, true},
		{"/a/123/z", "/z", 2, true},
		{"/a/123/z", "", 3, true},

		{"/a%2f123%2fz", "/a%2f123%2fz", 0, false},
		{"/a%2f123%2fz", "/123/z", 1, true},
		{"/a%2f123%2fz", "/z", 2, true},
		{"/a%2f123%2fz", "", 3, true},
	}

	for _, c := range cases {
		var err error
		a := NewStubHandler()
		req := httptest.NewRequest("", "/", nil)
		req.URL, err = url.ParseRequestURI(c.url)
		if err != nil {
			t.Fatal("Invalid test using", c.url)
		}
		w := httptest.NewRecorder()

		hf := StripLeadingSegments(c.strip, a)
		hf.ServeHTTP(w, req)

		if !c.mustDiffer && a.CapturedRequest != req {
			t.Errorf("Expected same request for %s", c.url)
		}
		if c.mustDiffer && a.CapturedRequest == req {
			t.Errorf("Expected different request for %s", c.url)
		}
		if a.CapturedRequest.URL.String() != c.remaining {
			t.Errorf("Expected %s; got %s", c.remaining, a.CapturedRequest.URL.String())
		}
		if c.mustDiffer && a.CapturedRequest.URL.RawPath != "" {
			t.Errorf("For %s got %s", c.remaining, a.CapturedRequest.URL.RawPath)
		}
	}
}
