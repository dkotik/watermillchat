package hypermedia

import (
	"bytes"
	"context"
	_ "embed" // for error.html
	"errors"
	"fmt"
	"io"
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

type RenderableError interface {
	Error
	Renderable
	Describe() (title, description *i18n.LocalizeConfig)
}

type ErrorHandler interface {
	HandlerError(http.ResponseWriter, *http.Request, error)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandlerError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

var DefaultErrorHandler = NewErrorPageHandler(
	i18n.NewBundle(DefaultLanguage),
	Head{},
	[]RenderableError{
		ErrNotFound,
		ErrInternalServerError,
	})

var PlainTextErrorHandler = ErrorHandlerFunc(
	func(w http.ResponseWriter, r *http.Request, err error) {
		statusCode := http.StatusInternalServerError
		var errorWithStatusCode Error
		if errors.As(err, &errorWithStatusCode) {
			statusCode = errorWithStatusCode.HyperTextStatusCode()
		}
		http.Error(w, err.Error(), statusCode)
	},
)

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
}

func (h prerenderedErrorHandler) HandlerError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	var errorWithStatusCode Error
	if errors.As(err, &errorWithStatusCode) {
		statusCode = errorWithStatusCode.HyperTextStatusCode()
	}
	w.WriteHeader(statusCode)
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
	Renderer   Renderable
}

// NewErrorPageHandler creates a pre-rendered localized error
// reponses based on [Error.HyperTextStatusCode]. Falls back on the
// renderer associated with [http.StatusInternalServerError].
// If that renderer is missing, creates one.
func NewErrorPageHandler(
	bundle *i18n.Bundle,
	head Head,
	renderableErrors []RenderableError,
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

	r := prerenderedErrorHandler{
		Pages: make(map[codeToLocalizedErrorPageKey][]byte),
	}

	ctx := context.Background()
	for _, tag := range languages {
		for _, renderable := range renderableErrors {
			localizer := i18n.NewLocalizer(bundle, tag.String())
			b := &bytes.Buffer{}

			head.Title, head.Description = renderable.Describe()
			// TODO: inject style/error.css is head.Stylesheets is len=0, and below
			if err := NewPageRenderer(head)(renderable).Render(ctx, b, localizer); err != nil {
				panic(fmt.Errorf("unable to render error: %w", err))
			}
			statusCode := renderable.HyperTextStatusCode()
			r.Pages[codeToLocalizedErrorPageKey{
				HyperTextStatusCode: statusCode,
				Language:            tag,
			}] = b.Bytes()

			if statusCode == http.StatusInternalServerError {
				r.InternalServerError = append(r.InternalServerError, prerenderedPage{
					Language:  tag,
					HyperText: b.Bytes(),
				})
			}
		}
	}

	r.Matcher = language.NewMatcher(languages)
	bottom, ok := r.Pages[codeToLocalizedErrorPageKey{
		HyperTextStatusCode: http.StatusInternalServerError,
		Language:            languages[0],
	}]
	if !ok {
		b := &bytes.Buffer{}
		if err := NewPageRenderer(head)(ErrInternalServerError).Render(
			ctx, b, i18n.NewLocalizer(bundle, languages[0].String())); err != nil {
			panic(err)
		}
	}
	r.Bottom = bottom
	return r
}

type errorTemplateValues struct {
	Title                 string
	Description           string
	StatusCodeDescription string
}

var (
	//go:embed error.html
	errorTemplateSource string
	errorTemplate       = template.Must(template.New("error").Parse(errorTemplateSource))

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

type renderableError int

func (e renderableError) Error() string {
	return http.StatusText(int(e))
}

func (e renderableError) HyperTextStatusCode() int {
	return int(e)
}

func (e renderableError) Describe() (title, description *i18n.LocalizeConfig) {
	switch e {
	case http.StatusNotFound:
		return &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.NotFoundError.Title",
					Other: "Requested Page Does Not Exist",
				},
			}, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.NotFoundError.Description",
					Other: "There is no content for this link.",
				},
			}
	case http.StatusForbidden:
		return &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.ForbiddenError.Title",
					Other: "Access Denied",
				},
			}, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.ForbiddenError.Description",
					Other: "You are not allowed to access this page.",
				},
			}
	default: // http.StatusInternalServerError
		return &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.InternalServerErrorError.Title",
					Other: "Internal Server Error",
				},
			}, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "hypermedia.InternalServerErrorError.Description",
					Other: "Server is unable to complete this request.",
				},
			}
	}
}

func (e renderableError) Render(ctx context.Context, w io.Writer, l *i18n.Localizer) error {
	title, description := e.Describe()
	titleText, err := l.Localize(title)
	if err != nil {
		return err
	}
	descriptionText, err := l.Localize(description)
	if err != nil {
		return err
	}
	scDescription, err := l.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "hypermedia.error.StatusCodeDescription",
			Other: "System assigned page error code: #{{.StatusCode}}.",
		},
		TemplateData: map[string]interface{}{
			"StatusCode": int(e),
		},
	})

	return errorTemplate.Execute(w, errorTemplateValues{
		Title:                 titleText,
		Description:           descriptionText,
		StatusCodeDescription: scDescription,
	})
}

const (
	ErrNotFound            = renderableError(http.StatusNotFound)
	ErrForbidden           = renderableError(http.StatusForbidden)
	ErrInternalServerError = renderableError(http.StatusInternalServerError)
)
