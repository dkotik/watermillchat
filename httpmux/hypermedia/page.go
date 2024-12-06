package hypermedia

import (
	"bytes"
	_ "embed" // for page template
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
)

var (
	//go:embed page.html
	pageTemplate string

	//go:embed style/page.css
	pageStyle []byte
)

type PageValues struct {
	Locale      string
	Title       string
	Description string
	Image       string
	FavIconPNG  string
	Scripts     []string
	Stylesheets []string
	Body        string
}

func NewPageRenderer(
	h Handler,
	errorHandler ErrorHandler,
	errorLogger *slog.Logger,
	v PageValues,
) http.HandlerFunc {
	if v.Body != "" {
		panic("page renderer body must be empty as it will be replaced by output from the handler")
	}
	const bodySplitInjection = `<<<<<<<<<<<<<>>>>>>>>>>>>>`
	v.Body = bodySplitInjection

	b := &bytes.Buffer{}
	tmpl := template.Must(template.New("page").Parse(pageTemplate))

	if err := tmpl.Execute(b, v); err != nil {
		panic(fmt.Errorf("unable to execute page.html template: %w", err))
	}
	head, tail, ok := bytes.Cut(b.Bytes(), []byte(bodySplitInjection))
	if !ok {
		panic("unable to cut rendered page.html template at body")
	}
	b.Reset()

	if errorHandler == nil {
		v.Title = "Internal Server Error"
		v.Description = "Your requested failed to complete because of a server error."
		if err := tmpl.Execute(b, v); err != nil {
			panic(fmt.Errorf("unable to execute page.html template: %w", err))
		}
		errorHead, _, ok := bytes.Cut(b.Bytes(), []byte(bodySplitInjection))
		if !ok {
			panic("unable to cut rendered page.html template at body")
		}
		b.Reset()
		errorHandler = NewErrorHandler(errorHead, tail, errorLogger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(b, r); err != nil {
			errorHandler(w, r, err)
			return
		}
		_, _ = io.Copy(w, bytes.NewReader(head))
		_, _ = io.Copy(w, b)
		_, _ = io.Copy(w, bytes.NewReader(tail))
		b.Reset()
	}
}

func PageStyleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	_, _ = io.Copy(w, bytes.NewReader(pageStyle))
}
