// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"fmt"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"
)

func TestParams(t *testing.T) {
	g := NewGomegaWithT(t)
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		g.Expect(ps.ByName(ps[i].Key)).To(Equal(ps[i].Value))
	}
	g.Expect(ps.ByName("noKey")).To(Equal(""))
}

func TestRouter_nested_handler_with_params_at_both_levels(t *testing.T) {
	g := NewGomegaWithT(t)
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
		g.Expect(ps).To(Equal(want))
	}))

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/top/rank/user/gopher", nil)
	r2.ServeHTTP(w, req)

	g.Expect(routed).To(BeTrue())
}

type handlerStruct struct {
	handled *int
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handled++
}

func TestRouterAPI(t *testing.T) {
	g := NewGomegaWithT(t)
	var get, head, options, post, put, patch, delete, handler, handlerFunc int

	httpHandler := handlerStruct{&handler}

	router := New()
	router.GET("/GET", func(w http.ResponseWriter, r *http.Request, _ Params) {
		get++
	})
	router.HEAD("/HEAD", func(w http.ResponseWriter, r *http.Request, _ Params) {
		head++
	})
	router.OPTIONS("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		options++
	})
	router.POST("/POST", func(w http.ResponseWriter, r *http.Request, _ Params) {
		post++
	})
	router.PUT("/PUT", func(w http.ResponseWriter, r *http.Request, _ Params) {
		put++
	})
	router.PATCH("/PATCH", func(w http.ResponseWriter, r *http.Request, _ Params) {
		patch++
	})
	router.DELETE("/DELETE", func(w http.ResponseWriter, r *http.Request, _ Params) {
		delete++
	})
	router.Handler(http.MethodGet, "/Handler", httpHandler)
	router.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc++
	})

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/GET", nil)
	router.ServeHTTP(w, r)
	g.Expect(get).To(Equal(1))

	r, _ = http.NewRequest(http.MethodHead, "/HEAD", nil)
	router.ServeHTTP(w, r)
	g.Expect(head).To(Equal(1))

	r, _ = http.NewRequest(http.MethodOptions, "/", nil)
	router.ServeHTTP(w, r)
	g.Expect(options).To(Equal(1))

	r, _ = http.NewRequest(http.MethodPost, "/POST", nil)
	router.ServeHTTP(w, r)
	g.Expect(post).To(Equal(1))

	r, _ = http.NewRequest(http.MethodPut, "/PUT", nil)
	router.ServeHTTP(w, r)
	g.Expect(put).To(Equal(1))

	r, _ = http.NewRequest(http.MethodPatch, "/PATCH", nil)
	router.ServeHTTP(w, r)
	g.Expect(patch).To(Equal(1))

	r, _ = http.NewRequest(http.MethodDelete, "/DELETE", nil)
	router.ServeHTTP(w, r)
	g.Expect(delete).To(Equal(1))

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	router.ServeHTTP(w, r)
	g.Expect(handler).To(Equal(1))

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	g.Expect(handlerFunc).To(Equal(1))
}

func TestRouter_API_using_implicit_HEAD(t *testing.T) {
	g := NewGomegaWithT(t)
	var get int

	router := New()
	router.GET("/GET", func(w http.ResponseWriter, r *http.Request, _ Params) {
		get++
	})

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodHead, "/GET", nil)
	router.ServeHTTP(w, r)
	g.Expect(get).To(Equal(1))
}

func TestRouter_HandleAll(t *testing.T) {
	g := NewGomegaWithT(t)
	var saw = make(map[string]int)

	router := New()
	router.HandleAll("/", func(w http.ResponseWriter, r *http.Request, _ Params) {
		saw[r.Method]++
	})

	w := httptest.NewRecorder()

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions} {
		r, _ := http.NewRequest(method, "/", nil)
		router.ServeHTTP(w, r)
	}

	g.Expect(saw).To(Equal(map[string]int{
		http.MethodHead: 1, http.MethodGet: 1, http.MethodPut: 1, http.MethodPost: 1, http.MethodDelete: 1, http.MethodPatch: 1, http.MethodOptions: 1,
	}))
}

func TestRouter_HandlerAll(t *testing.T) {
	g := NewGomegaWithT(t)
	var saw = make(map[string]int)

	router := New()
	router.HandlerAll("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw[r.Method]++
	}))

	w := httptest.NewRecorder()

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions} {
		r, _ := http.NewRequest(method, "/", nil)
		router.ServeHTTP(w, r)
	}

	g.Expect(saw).To(Equal(map[string]int{
		http.MethodHead: 1, http.MethodGet: 1, http.MethodPut: 1, http.MethodPost: 1, http.MethodDelete: 1, http.MethodPatch: 1, http.MethodOptions: 1,
	}))
}

func TestRouter_Root(t *testing.T) {
	g := NewGomegaWithT(t)
	router := New()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})
	g.Expect(recv).NotTo(BeNil())
}

func TestRouterInvalidInput(t *testing.T) {
	g := NewGomegaWithT(t)
	router := New()

	handle := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	recv := catchPanic(func() {
		router.Handle("", "/", handle)
	})
	g.Expect(recv).NotTo(BeNil())

	recv = catchPanic(func() {
		router.GET("", handle)
	})
	g.Expect(recv).NotTo(BeNil())

	recv = catchPanic(func() {
		router.GET("noSlashRoot", handle)
	})
	g.Expect(recv).NotTo(BeNil())

	recv = catchPanic(func() {
		router.GET("/", nil)
	})
	g.Expect(recv).NotTo(BeNil())
}
func TestRouter_Chaining(t *testing.T) {
	g := NewGomegaWithT(t)
	router1 := New()
	router2 := New()
	router1.NotFound = router2

	fooHit := 0
	router1.POST("/foo", func(w http.ResponseWriter, req *http.Request, _ Params) {
		fooHit++
		w.WriteHeader(http.StatusOK)
	})

	barHit := 0
	router2.POST("/bar", func(w http.ResponseWriter, req *http.Request, _ Params) {
		barHit++
		w.WriteHeader(http.StatusOK)
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	g.Expect(fooHit).To(Equal(1))
	g.Expect(w.Code).To(Equal(http.StatusOK))

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	g.Expect(barHit).To(Equal(1))
	g.Expect(w.Code).To(Equal(http.StatusOK))

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNotFound))
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
	g := NewGomegaWithT(t)
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest(http.MethodOptions, "*", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusOK))
	g.Expect(w.Header().Get("Allow")).To(Equal("OPTIONS, POST"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusOK))
	g.Expect(w.Header().Get("Allow")).To(Equal("OPTIONS, POST"))

	r, _ = http.NewRequest(http.MethodOptions, "/doesnotexist", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNotFound))

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
	g.Expect(w.Code).To(Equal(http.StatusNoContent))
	g.Expect(w.Header().Get("Allow")).To(Equal("GET, OPTIONS, POST"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNoContent))
	g.Expect(w.Header().Get("Allow")).To(Equal("GET, OPTIONS, POST"))

	// custom handler
	custom := 0
	router.OPTIONS("/path", func(w http.ResponseWriter, r *http.Request, _ Params) {
		custom++
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNoContent))
	g.Expect(w.Header().Get("Allow")).To(Equal("GET, OPTIONS, POST"))
	g.Expect(custom).To(Equal(0))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusOK))
	g.Expect(custom).To(Equal(1))
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	g := NewGomegaWithT(t)
	handlerFunc := func(_ http.ResponseWriter, _ *http.Request, _ Params) {}

	router := New()
	router.POST("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusMethodNotAllowed))
	g.Expect(w.Header().Get("Allow")).To(Equal("OPTIONS, POST"))

	// add another method
	router.DELETE("/path", handlerFunc)
	router.OPTIONS("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusMethodNotAllowed))
	g.Expect(w.Header().Get("Allow")).To(Equal("DELETE, OPTIONS, POST"))

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	router.ServeHTTP(w, r)
	g.Expect(w.Body.String()).To(Equal(responseText))
	g.Expect(w.Code).To(Equal(http.StatusTeapot))
	g.Expect(w.Header().Get("Allow")).To(Equal("DELETE, OPTIONS, POST"))
}

func TestRouter_NotFound(t *testing.T) {
	g := NewGomegaWithT(t)
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
	for i, tr := range testRoutes {
		r, _ := http.NewRequest(http.MethodGet, tr.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		g.Expect(w.Code).To(Equal(tr.code))
		if w.Code != http.StatusNotFound {
			g.Expect(w.Header().Get("Location")).To(Equal(tr.location), fmt.Sprintf("%d %s", i, tr.route))
		}
	}

	// Test custom not found handler
	notFound := 0
	router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		notFound++
	})
	r, _ := http.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNotFound))
	g.Expect(notFound).To(Equal(1))

	// Test other method than GET (want 308 instead of 301)
	router.PATCH("/path", handlerFunc)
	r, _ = http.NewRequest(http.MethodPatch, "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusPermanentRedirect))
	g.Expect(w.Header().Get("Location")).To(Equal("/path"))

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", handlerFunc)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	g.Expect(w.Code).To(Equal(http.StatusNotFound))
}

func TestRouter_PanicHandler(t *testing.T) {
	g := NewGomegaWithT(t)
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
		g.Expect(recover()).To(BeNil())
	}()

	router.ServeHTTP(w, req)

	g.Expect(panicHandled).To(BeTrue())
}

func TestRouter_Lookup(t *testing.T) {
	g := NewGomegaWithT(t)
	routedCount := 0
	wantHandle := func(_ http.ResponseWriter, _ *http.Request, _ Params) {
		routedCount++
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := New()

	// try empty router first
	handle, _, tsr := router.Lookup(http.MethodGet, "/nope")
	g.Expect(handle).To(BeNil())
	g.Expect(tsr).To(BeFalse())

	// insert route and try again
	router.GET("/user/:name", wantHandle)

	handle, params, tsr := router.Lookup(http.MethodGet, "/user/gopher")
	g.Expect(handle).NotTo(BeNil())

	handle(nil, nil, nil)
	g.Expect(routedCount).To(Equal(1))
	g.Expect(params).To(Equal(wantParams))

	routedCount = 0

	// route without param
	router.GET("/user", wantHandle)
	handle, params, _ = router.Lookup(http.MethodGet, "/user")
	g.Expect(handle).NotTo(BeNil())

	handle(nil, nil, nil)
	g.Expect(routedCount).NotTo(Equal(0))

	g.Expect(params).To(BeNil())

	handle, _, tsr = router.Lookup(http.MethodGet, "/user/gopher/")
	g.Expect(handle).To(BeNil())
	g.Expect(tsr).To(BeTrue())

	handle, _, tsr = router.Lookup(http.MethodGet, "/nope")
	g.Expect(handle).To(BeNil())
	g.Expect(tsr).To(BeFalse())
}

func TestRouter_ListPaths(t *testing.T) {
	g := NewGomegaWithT(t)
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
	g.Expect(len(all)).To(Equal(len(AllMethods)))

	for _, method := range AllMethods {
		actual := all[method]
		extra := "/" + method
		expected := append(routes, extra)
		sort.Strings(expected)
		g.Expect(actual).To(Equal(expected))
	}

	actual := router.ListPaths(http.MethodGet)
	extra := "/GET"
	expected := append(routes, extra)
	sort.Strings(expected)
	g.Expect(actual[http.MethodGet]).To(Equal(expected))
}

//-------------------------------------------------------------------------------------------------

func TestRouter_SubRouter_panics(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []string{"/noFilepath", "/foo*/*", "/foo*/*filepath", "/foo/"}

	for _, c := range cases {
		router := New()

		recv := catchPanic(func() {
			router.SubRouter(c, NewStubHandler())
		})
		g.Expect(recv).NotTo(BeNil(), c)
	}
}

func TestRouter_ServeFiles_supported_methods(t *testing.T) {
	g := NewGomegaWithT(t)
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
		g.Expect(recv).NotTo(BeNil(), c)

		router.ServeFiles("/*filepath", mfs)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(c.method, "/favicon.ico", nil)
		router.ServeHTTP(w, r)
		g.Expect(mfs.opened).To(Equal(1), c.method)
		g.Expect(w.Body.Len()).To(Equal(c.expectedLength), c.method)
	}
}

func TestRouter_ServeFiles_unsupported_methods(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []string{http.MethodPut, http.MethodPost, http.MethodDelete}

	for _, method := range cases {
		router := New()
		mfs := &mockFileSystem{}

		router.ServeFiles("/*filepath", mfs)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(method, "/favicon.ico", nil)
		router.ServeHTTP(w, r)
		g.Expect(mfs.opened).To(BeZero(), method)
		g.Expect(w.Code).To(Equal(405), method)
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
