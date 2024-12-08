package hypermedia

import (
	"bytes"
	"context"
	_ "embed" // for error.html
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Error interface {
	error
	HyperTextStatusCode() int
}

type ErrorHandler interface {
	HandlerError(http.ResponseWriter, *http.Request, error)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandlerError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

type codeToLocalizedErrorPageKey struct {
	HyperTextStatusCode int
	Language            language.Tag
}

type prerenderedErrorHandler struct {
	Matcher             language.Matcher
	Pages               map[codeToLocalizedErrorPageKey][]byte
	InternalServerError []prerenderedPage

	// Bottom is written with [http.StatusInternalServerError]
	// when nothing else matched.
	Bottom []byte
	Logger *slog.Logger
}

func (h prerenderedErrorHandler) HandlerError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	var errorWithStatusCode Error
	if errors.As(err, &errorWithStatusCode) {
		statusCode = errorWithStatusCode.HyperTextStatusCode()
	}
	w.WriteHeader(statusCode)

	defer h.Logger.ErrorContext(
		r.Context(),
		"request failed to complete",
		slog.Any("error", err),
		slog.Int("status_code", statusCode),
		slog.String("host", r.URL.Hostname()),
		slog.String("path", r.URL.Path),
		slog.String("client_address", r.RemoteAddr),
	)

	tags, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	language, _, _ := h.Matcher.Match(tags...)

	hyperText, ok := h.Pages[codeToLocalizedErrorPageKey{
		HyperTextStatusCode: statusCode,
		Language:            language,
	}]
	if ok {
		// matched combination of status code and language
		_, _ = io.Copy(w, bytes.NewReader(hyperText))
		return
	}

	for _, page := range h.InternalServerError {
		if page.Language == language {
			// matched known language
			_, _ = io.Copy(w, bytes.NewReader(page.HyperText))
			return
		}
	}
	_, _ = io.Copy(w, bytes.NewReader(h.Bottom))
}

type ErrorRenderer struct {
	StatusCode int
	Renderer   Renderer
}

// NewErrorPageHandler creates a pre-rendered localized error
// reponses based on [Error.GetStatusCode]. Falls back on the
// renderer associated with [http.StatusInternalServerError].
// If that renderer is missing, creates one.
func NewErrorPageHandler(
	bundle *i18n.Bundle,
	renderers []ErrorRenderer,
	logger *slog.Logger,
	prioritizeLanguages ...string,
) ErrorHandler {
	var languages []language.Tag
	if bundle == nil {
		languages = []language.Tag{DefaultLanguage}
	} else {
		languages = bundle.LanguageTags()
		if len(languages) > 1 {
			if err := SortLanguageTags(languages, strings.Join(prioritizeLanguages, ",")); err != nil {
				panic(err)
			}
		}
	}
	defaultRenderer := func() Renderer {
		for _, r := range renderers {
			if r.StatusCode == http.StatusInternalServerError {
				return r.Renderer
			}
		}
		bottom := RendererFunc(
			func(_ context.Context, w io.Writer, l *i18n.Localizer) error {
				// TODO: prettify
				message, err := l.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "hypermedia.InternalServerError",
						Other: "Internal Server Error",
					},
				})
				if err != nil {
					return err
				}
				_, err = io.WriteString(w, message)
				return err
			},
		)
		renderers = append(renderers, ErrorRenderer{
			StatusCode: http.StatusInternalServerError,
			Renderer:   bottom,
		})
		return bottom
	}()
	if logger == nil {
		logger = slog.Default()
	}

	r := prerenderedErrorHandler{
		Logger: logger,
	}
	for _, renderer := range renderers {
		for _, tag := range languages {
			localizer := i18n.NewLocalizer(bundle, tag.String())
			b := &bytes.Buffer{}
			if err := renderer.Renderer.Render(context.TODO(), b, localizer); err != nil {
				panic(fmt.Errorf("unable to render error: %w", err))
			}
			r.Pages[codeToLocalizedErrorPageKey{
				HyperTextStatusCode: renderer.StatusCode,
				Language:            tag,
			}] = b.Bytes()
		}
	}

	for _, tag := range languages {
		localizer := i18n.NewLocalizer(bundle, tag.String())
		b := &bytes.Buffer{}
		if err := defaultRenderer.Render(context.TODO(), b, localizer); err != nil {
			panic(fmt.Errorf("unable to render error: %w", err))
		}
		r.InternalServerError = append(r.InternalServerError, prerenderedPage{
			Language:  tag,
			HyperText: b.Bytes(),
		})
	}

	r.Matcher = language.NewMatcher(languages)
	r.Bottom = r.Pages[codeToLocalizedErrorPageKey{
		HyperTextStatusCode: http.StatusInternalServerError,
		Language:            languages[0],
	}]
	return r
}

type errorTemplateValues struct {
	Title       string
	Description string
}

var (
	//go:embed error.html
	errorTemplateSource string
	errorTemplate       = template.Must(template.New("error").Parse(pageTemplateSource))

	ErrNotFoundRenderer = RendererFunc(
		func(ctx context.Context, w io.Writer, l *i18n.Localizer) error {
			title, err := l.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.NotFoundError.Title",
					Other: "Requested Page Does Not Exist",
				},
			})
			if err != nil {
				return err
			}
			description, err := l.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.NotFoundError.Description",
					Other: "There is no content for this link.",
				},
			})
			if err != nil {
				return err
			}

			return errorTemplate.Execute(w, errorTemplateValues{
				Title:       title,
				Description: description,
			})
		},
	)
)
