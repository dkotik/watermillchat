/*
Package hypermedia provides templates and [http.Handler]s for rendering chat sessions.
*/
package hypermedia

import (
	_ "embed" // for including assets

	"bytes"
	"io"
	"net/http"
)

type (
	Handler func(main io.Writer, r *http.Request) error
)

var (
	//go:embed icon/favicon.png
	faviconPNG []byte

	//go:embed script/datastar.js
	datastarSource []byte

	//go:embed script/datastar.js.map
	datastarSourceMap []byte
)

func FavIconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	_, _ = io.Copy(w, bytes.NewReader(faviconPNG))
}

func DatastarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	_, _ = io.Copy(w, bytes.NewReader(datastarSource))
}

func DatastarMapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.Copy(w, bytes.NewReader(datastarSourceMap))
}

func NewAsset(contentType string, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		_, _ = io.Copy(w, bytes.NewReader(body))
	}
}
