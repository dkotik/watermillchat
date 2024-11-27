/*
Package datastar provides <data-star.dev> view rendering for
[watermillchat.Server].
*/
package datastar

import (
	"bytes"
	_ "embed" // for bundindling datastar
	"io"
	"net/http"
)

//go:embed datastar.js
var scriptSource []byte

//go:embed datastar.js.map
var scriptSourceMap []byte

func SourceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	_, _ = io.Copy(w, bytes.NewReader(scriptSource))
}

func SourceMapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.Copy(w, bytes.NewReader(scriptSourceMap))
}
