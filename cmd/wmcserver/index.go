package main

import (
	_ "embed" // for bundling index.html
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	datastarLibrary "github.com/starfederation/datastar/code/go/sdk"
)

//go:embed index.html
var indexTemplate string

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, strings.NewReader(indexTemplate))
}

func EventHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastarLibrary.NewSSE(w, r)

	sse.MergeFragments(
		`<title>Updated Title</title>`,
		datastarLibrary.WithSelector("title"),
	)

	sse.MergeSignals([]byte(`{response: '', answer: 'bread'}`))

	for i := range 10 {
		time.Sleep(time.Second)
		sse.MergeFragments(
			fmt.Sprintf(`<div id="question" data-view-transition="foo">%d ? %d ?</div>`,
				i,
				time.Now().Unix()),
			datastarLibrary.WithMergeMode(datastarLibrary.FragmentMergeModeAppend),
			datastarLibrary.WithSettleDuration(time.Second*2),
			datastarLibrary.WithViewTransitions(),
		)
		sse.MergeSignals([]byte(
			fmt.Sprintf(`{title: '%d - %d'}`, i,
				time.Now().Unix())))
	}
}
