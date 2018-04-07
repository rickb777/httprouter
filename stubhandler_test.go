// Copyright 2018 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import "net/http"

// Provides a test stub to replace handlers in HTTP wiring tests.
// Normally, this will be used in conjunction with httptest.ResponseRecorder.
type StubHandler struct {
	CapturedRequest *http.Request
	CapturedWriter  http.ResponseWriter
	testHandler     http.HandlerFunc
}

var _ http.Handler = &StubHandler{}

// NewStubHandler creates a new stub handler.
func NewStubHandler() *StubHandler {
	return &StubHandler{}
}

// OnServe allows the stub to act as a proxy for another handler, supplied here.
// This turns the 'stub' into a 'spy'.
func (h *StubHandler) OnServe(testHandler http.HandlerFunc) {
	h.testHandler = testHandler
}

// ServeHTTP is http.Handler.
func (h *StubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.CapturedRequest = r
	h.CapturedWriter = w

	if h.testHandler != nil {
		h.testHandler(w, r)
	}
}
