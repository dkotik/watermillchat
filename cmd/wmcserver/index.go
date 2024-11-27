package main

import (
	_ "embed" // for bundling index.html
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	datastar "github.com/starfederation/datastar/code/go/sdk"
)

//go:embed index.html
var indexTemplate string

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, strings.NewReader(indexTemplate))
}

func EventHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Merges HTML fragments into the DOM.
	sse.MergeFragments(
		`<div id="question">What do you put in a toaster?</div>`)

	// Merges signals into the store.
	sse.MergeSignals([]byte(`{response: '', answer: 'bread'}`))

	for i := range 10 {
		time.Sleep(time.Second)
		sse.MergeFragments(
			fmt.Sprintf(`<div id="question">%d ? %d ?</div>`,
				i,
				time.Now().Unix()))
		sse.MergeSignals([]byte(
			fmt.Sprintf(`{title: '%d - %d'}`, i,
				time.Now().Unix())))
	}
}
