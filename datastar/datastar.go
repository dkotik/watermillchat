/*
Package datastar provides <data-star.dev> view rendering for
[watermillchat.Server].
*/
package datastar

import (
	"bytes"
	_ "embed" // for bundindling datastar
	"html/template"
	"io"
	"log/slog"
	"net/http"

	datastar "github.com/starfederation/datastar/code/go/sdk"
)

//go:embed datastar.js
var scriptSource []byte

//go:embed datastar.js.map
var scriptSourceMap []byte

var errorTemplate = template.Must(template.New("error").Parse(`<p class="error">{{.}}</p>`))

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			sse := datastar.NewSSE(w, r)

			b := &bytes.Buffer{}
			if err = errorTemplate.Execute(b, err.Error()); err != nil {
				panic(err)
			}
			sse.MergeFragments(
				b.String(),
				datastar.WithSelector(".error"),
			)
		}
	}
}

func SourceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	_, _ = io.Copy(w, bytes.NewReader(scriptSource))
}

func SourceMapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.Copy(w, bytes.NewReader(scriptSourceMap))
}
