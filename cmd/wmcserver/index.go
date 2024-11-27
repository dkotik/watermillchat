package main

import (
	_ "embed" // for bundling index.html
	"io"
	"net/http"
	"strings"
)

//go:embed index.html
var indexTemplate string

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, strings.NewReader(indexTemplate))
}
