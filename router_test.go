// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.ByName(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.ByName("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestRouter_nested_handler_with_params_at_both_levels(t *testing.T) {
	r1 := New()
	r2 := New()

	routed := false

	r2.SubRouter("/top/:top/*", r1)
	r1.Handler(http.MethodGet, "/user/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps := ParamsFromContext(r.Context())
		routed = true
		want := Params{
			Param{"top", "rank"},
			Param{"filepath", "/user/gopher"},
			Param{"name", "gopher"},
		}
		if !reflect.DeepEqual(ps, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, ps)
		}
	}))

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/top/rank/user/gopher", nil)
	r2.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

type handlerStruct struct {
	handled *int
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handled++
}

func TestRouter_API_using_automatic_HEAD(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc int

	httpHandler := handlerStruct{&handler}

	router := New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		get++
	})
	router.HEAD("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		head++
	})
	router.OPTIONS("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		options++
	})
	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		post++
	})
	router.PUT("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		put++
	})
	router.PATCH("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		patch++
	})
	router.DELETE("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		delete++
	})
	router.Handler(http.MethodGet, "/Handler", httpHandler)
	router.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc++
	})

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, r)
	if get != 1 {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest(http.MethodHead, "/", nil)
	router.ServeHTTP(w, r)
	if head != 1 {
		t.Error("routing HEAD failed")
	}

	_, _, isHead := router.Lookup(http.MethodHead, "/")
	if isHead { // <-- N.B.
		t.Error("lookup HEAD failed")
	}

	r, _ = http.NewRequest(http.MethodOptions, "/", nil)
	router.ServeHTTP(w, r)
	if options != 1 {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest(http.MethodPost, "/", nil)
	router.ServeHTTP(w, r)
	if post != 1 {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest(http.MethodPut, "/", nil)
	router.ServeHTTP(w, r)
	if put != 1 {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest(http.MethodPatch, "/", nil)
	router.ServeHTTP(w, r)
	if patch != 1 {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest(http.MethodDelete, "/", nil)
	router.ServeHTTP(w, r)
	if delete != 1 {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	router.ServeHTTP(w, r)
	if handler != 1 {
		t.Error("routing Handler failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	if handlerFunc != 1 {
		t.Error("routing HandlerFunc failed")
	}
}

func TestRouter_API_using_specialised_HEAD(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc int

	httpHandler := handlerStruct{&handler}

	router := New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		get++
	})
	router.HEAD("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		head++
	})
	router.OPTIONS("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		options++
	})
	router.POST("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		post++
	})
	router.PUT("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		put++
	})
	router.PATCH("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		patch++
	})
	router.DELETE("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		delete++
	})
	router.Handler(http.MethodGet, "/Handler", httpHandler)
	router.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc++
	})

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, r)
	if get != 1 {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest(http.MethodHead, "/", nil)
	router.ServeHTTP(w, r)
	if head != 1 {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest(http.MethodOptions, "/", nil)
	router.ServeHTTP(w, r)
	if options != 1 {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest(http.MethodPost, "/", nil)
	router.ServeHTTP(w, r)
	if post != 1 {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest(http.MethodPut, "/", nil)
	router.ServeHTTP(w, r)
	if put != 1 {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest(http.MethodPatch, "/", nil)
	router.ServeHTTP(w, r)
	if patch != 1 {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest(http.MethodDelete, "/", nil)
	router.ServeHTTP(w, r)
	if delete != 1 {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	router.ServeHTTP(w, r)
	if handler != 1 {
		t.Error("routing Handler failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	if handlerFunc != 1 {
		t.Error("routing HandlerFunc failed")
	}
}

func TestRouter_HandleAll(t *testing.T) {
	var saw = make(map[string]struct{})

	router := New()
	router.HandleAll("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		saw[r.Method] = struct{}{}
	})

	w := httptest.NewRecorder()

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions} {
		r, _ := http.NewRequest(method, "/", nil)
		router.ServeHTTP(w, r)
		if _, ok := saw[method]; !ok {
			t.Errorf("routing %s failed", method)
		}
	}
}

func TestRouter_Root(t *testing.T) {
	router := New()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouter_Chaining(t *testing.T) {
	router1 := New()
	router2 := New()
	router1.NotFound = router2

	fooHit := false
	router1.POST("/foo", func(w http.ResponseWriter, req *http.Request, _ Params) {
		fooHit = true
		w.WriteHeader(http.StatusOK)
	})

	barHit := false
	router2.POST("/bar", func(w http.ResponseWriter, req *http.Request, _ Params) {
		barHit = true
		w.WriteHeader(http.StatusOK)
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func BenchmarkAllowed(b *testing.B) {
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.POST("/path", handlerFunc)
	router.GET("/path", handlerFunc)

	b.Run("Global", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("*", http.MethodOptions)
		}
	})
	b.Run("Path", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("/path", http.MethodOptions)
		}
	})
}
func TestRouter_OPTIONS(t *testing.T) {
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest(http.MethodOptions, "*", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	r, _ = http.NewRequest(http.MethodOptions, "/doesnotexist", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// add another method
	router.GET("/path", handlerFunc)

	// set a global OPTIONS handler
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	router.OPTIONS("/path", func(w http.ResponseWriter, r *http.Request, _ Params) {
		custom = true
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}
	if custom {
		t.Error("custom handler called on *")
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}
	if !custom {
		t.Error("custom handler not called")
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.DELETE("/path", handlerFunc)
	router.OPTIONS("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	router.ServeHTTP(w, r)
	if got := w.Body.String(); !(got == responseText) {
		t.Errorf("unexpected response got %q want %q", got, responseText)
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("unexpected response code %d want %d", w.Code, http.StatusTeapot)
	}
	if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}
}

func TestRouter_NotFound(t *testing.T) {
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)
	router.GET("/", handlerFunc)

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"", http.StatusMovedPermanently, "/"},             // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		r, _ := http.NewRequest(http.MethodGet, tr.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if !(w.Code == tr.code && (w.Code == http.StatusNotFound || fmt.Sprint(w.Header().Get("Location")) == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header().Get("Location"))
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		notFound = true
	})
	r, _ := http.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test other method than GET (want 308 instead of 301)
	router.PATCH("/path", handlerFunc)
	r, _ = http.NewRequest(http.MethodPatch, "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusPermanentRedirect && fmt.Sprint(w.Header()) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.Code)
	}
}

func TestRouter_PanicHandler(t *testing.T) {
	router := New()
	panicHandled := false

	router.PanicHandler = func(rw http.ResponseWriter, r *http.Request, p interface{}) {
		panicHandled = true
	}

	router.Handle(http.MethodPut, "/user/:name", func(_ http.ResponseWriter, _ *http.Request, _ Params) {
		panic("oops!")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/user/gopher", nil)

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	router.ServeHTTP(w, req)

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

func TestRouter_Lookup(t *testing.T) {
	routedCount := 0
	wantHandle := func(_ http.ResponseWriter, _ *http.Request, _ Params) {
		routedCount++
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := New()

	// try empty router first
	handle, _, tsr := router.Lookup(http.MethodGet, "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation.")
	}

	// insert route and try again
	router.GET("/user/:name", wantHandle)

	handle, params, tsr := router.Lookup(http.MethodGet, "/user/gopher")
	if handle == nil {
		t.Fatal("Got no handle.")
	} else {
		handle(nil, nil, nil)
		if routedCount != 1 {
			t.Fatal("Routing failed.")
		}
	}

	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}
	routedCount = 0

	// route without param
	router.GET("/user", wantHandle)
	handle, params, _ = router.Lookup(http.MethodGet, "/user")
	if handle == nil {
		t.Fatal("Got no handle!")
	} else {
		handle(nil, nil, nil)
		if routedCount == 0 {
			t.Fatal("Routing failed!")
		}
	}
	if params != nil {
		t.Fatalf("Wrong parameter values: want %v, got %v", nil, params)
	}

	handle, _, tsr = router.Lookup(http.MethodGet, "/user/gopher/")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation.")
	}

	handle, _, tsr = router.Lookup(http.MethodGet, "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation.")
	}
}

func TestRouter_ListPaths(t *testing.T) {
	router := New()

	routes := []string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool/",
		"/src/*filepath",
		"/search/",
		"/search/:query",
		"/user_:name",
		"/user_:name/about",
		"/files/:dir/*filepath",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
	}

	for _, method := range AllMethods {
		for _, route := range routes {
			router.Handle(method, route, fakeHandler(method+" "+route))
		}
		extra := "/" + method
		router.Handle(method, extra, fakeHandler(method+" "+extra))
	}

	//printChildren(router.trees[GET], "")

	all := router.ListPaths("")
	if len(all) != len(AllMethods) {
		t.Errorf("Expected %d methods but got %d\n%v", len(AllMethods), len(all), all)
	}

	for _, method := range AllMethods {
		actual := all[method]
		extra := "/" + method
		expected := append(routes, extra)
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("\nExpected %v\nbut got  %v", expected, actual)
		}
	}

	actual := router.ListPaths(http.MethodGet)
	extra := "/GET"
	expected := append(routes, extra)
	sort.Strings(expected)
	if !reflect.DeepEqual(expected, actual[http.MethodGet]) {
		t.Errorf("\nExpected %v\nbut got  %v", expected, actual)
	}
}

//-------------------------------------------------------------------------------------------------

func TestRouter_SubRouter_panics(t *testing.T) {
	cases := []string{"/noFilepath", "/foo*/*", "/foo*/*filepath", "/foo/"}

	for _, c := range cases {
		router := New()

		recv := catchPanic(func() {
			router.SubRouter(c, NewStubHandler())
		})
		if recv == nil {
			t.Errorf("%s: registering path not ending with '*filepath' did not panic", c)
		}
	}
}

func TestRouter_ServeFiles_supported_methods(t *testing.T) {
	cases := []struct {
		method         string
		expectedLength int
	}{
		{method: http.MethodGet, expectedLength: 5},
		{method: http.MethodHead, expectedLength: 0},
	}

	for _, c := range cases {
		router := New()
		mfs := &mockFileSystem{}

		recv := catchPanic(func() {
			router.ServeFiles("/noFilepath", mfs)
		})
		if recv == nil {
			t.Fatal("registering path not ending with '*filepath' did not panic")
		}

		router.ServeFiles("/*filepath", mfs)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(c.method, "/favicon.ico", nil)
		router.ServeHTTP(w, r)
		if mfs.opened != 1 {
			t.Errorf("%s: serving file not opened", c.method)
		}
		if w.Body.Len() != c.expectedLength {
			t.Errorf("%s: serving file wrong length %d", c.method, w.Body.Len())
		}
	}
}

func TestRouter_ServeFiles_unsupported_methods(t *testing.T) {
	cases := []string{http.MethodPut, http.MethodPost, http.MethodDelete}

	for _, method := range cases {
		router := New()
		mfs := &mockFileSystem{}

		router.ServeFiles("/*filepath", mfs)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(method, "/favicon.ico", nil)
		router.ServeHTTP(w, r)
		if mfs.opened != 0 {
			t.Errorf("%s: serving file should not happen", method)
		}
		if w.Code != 405 {
			t.Errorf("%s: code was %d", method, w.Code)
		}
	}
}

//-------------------------------------------------------------------------------------------------

type mockFileSystem struct {
	opened int
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened++
	return mockFile{name: name}, nil
}

type mockFile struct {
	name string
}

func (m mockFile) Close() error {
	return nil
}

func (m mockFile) Read(p []byte) (n int, err error) {
	copy(p, "hello")
	return 5, io.EOF
}

func (m mockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m mockFile) Readdir(_ int) ([]os.FileInfo, error) {
	panic("implement me")
}

func (m mockFile) Stat() (os.FileInfo, error) {
	return mockInfo{name: m.name}, nil
}

type mockInfo struct {
	dir  bool
	name string
}

func (m mockInfo) Name() string {
	return m.name
}

func (m mockInfo) Size() int64 {
	return 5
}

func (m mockInfo) Mode() os.FileMode {
	if m.dir {
		return os.ModeDir
	}
	return 0
}

func (m mockInfo) ModTime() time.Time {
	return time.Time{}
}

func (m mockInfo) IsDir() bool {
	return m.dir
}

func (m mockInfo) Sys() interface{} {
	panic("implement me")
}
