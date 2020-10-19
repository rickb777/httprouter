package httprouter_test

import (
	"github.com/rickb777/httprouter/v3"
	"github.com/rickb777/servefiles/v3"
	"log"
	"net/http"
	"time"
)

func ExampleRouter_SubRouter_usingAnAssetHandler() {
	// This is a webserver using the asset handler provided by
	// github.com/rickb777/servefiles/v3, which has enhanced
	// HTTP expiry, cache control, compression etc.
	// 'Normal' bespoke handlers are included as needed.

	// where the assets are stored (replace as required)
	localPath := "./assets"

	// how long we allow user agents to cache assets
	// (this is in addition to conditional requests, see
	// RFC7234 https://tools.ietf.org/html/rfc7234#section-5.2.2.8)
	maxAge := time.Hour

	h := servefiles.NewAssetHandler(localPath).WithMaxAge(maxAge)

	router := httprouter.New()
	// ... add other routes / handlers as required
	router.SubRouter("/files/*", h, http.MethodGet, http.MethodHead)

	log.Fatal(http.ListenAndServe(":8080", router))
}
