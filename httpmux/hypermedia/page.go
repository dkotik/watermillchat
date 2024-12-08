package hypermedia

import (
	"bytes"
	"context"
	_ "embed" // for page template
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const PageContentType = "text/html; charset=utf-8"

var (
	//go:embed style/page.css
	pageStyle []byte

	//go:embed page.html
	pageTemplateSource string
	pageTemplate       = template.Must(template.New("page").Parse(pageTemplateSource))
)

type Renderer interface {
	Render(context.Context, io.Writer, *i18n.Localizer) error
}

type RendererFunc func(context.Context, io.Writer, *i18n.Localizer) error

func (f RendererFunc) Render(ctx context.Context, w io.Writer, l *i18n.Localizer) error {
	return f(ctx, w, l)
}

type prerenderedPage struct {
	Language  language.Tag
	HyperText []byte
}

func NewPage(ctx context.Context, r Renderer, b *i18n.Bundle, prioritizeLanguages ...string) http.Handler {
	languages := b.LanguageTags()
	if len(prioritizeLanguages) > 0 {
		if err := SortLanguageTags(languages, strings.Join(prioritizeLanguages, ",")); err != nil {
			panic(err)
		}
	}

	if len(languages) == 1 {
		buffer := &bytes.Buffer{}
		if err := r.Render(ctx, buffer, i18n.NewLocalizer(b,
			append([]string{languages[0].String()}, prioritizeLanguages...)...,
		)); err != nil {
			panic(err)
		}
		return NewAsset(PageContentType, buffer.Bytes())
	}

	matcher := language.NewMatcher(languages)
	pageSet := make([]prerenderedPage, len(languages))
	for i, tag := range languages {
		buffer := &bytes.Buffer{}
		if err := r.Render(ctx, buffer, i18n.NewLocalizer(b,
			append([]string{tag.String()}, prioritizeLanguages...)...,
		)); err != nil {
			panic(err)
		}
		pageSet[i] = prerenderedPage{
			Language:  tag,
			HyperText: buffer.Bytes(),
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// languages supported by this service:
		// matcher := language.NewMatcher([]language.Tag{
		// 	language.English, language.Dutch, language.German,
		// })
		// 	lang, _ := r.Cookie("lang")
		// 	tag, _ := language.MatchStrings(matcher, lang.String(), r.Header.Get("Accept-Language"))
		tag, _ := language.MatchStrings(matcher, r.Header.Get("Accept-Language"))
		for _, page := range pageSet {
			if page.Language == tag {
				w.Header().Set("Content-Type", PageContentType)
				_, _ = io.Copy(w, bytes.NewReader(page.HyperText))
			}
		}
		panic("no locale tag matched")
	})
}

type Head struct {
	Title       *i18n.LocalizeConfig
	Description *i18n.LocalizeConfig
	Image       string
	FavIconPNG  string
	Scripts     []string
	Stylesheets []string
}

type pageTemplateValues struct {
	Language    string
	Title       string
	Description string
	Image       string
	FavIconPNG  string
	Scripts     []string
	Stylesheets []string
	Main        template.HTML
}

func NewPageRenderer(
	head Head,
	body Renderer,
) Renderer {
	return RendererFunc(func(
		ctx context.Context,
		w io.Writer,
		l *i18n.Localizer,
	) (err error) {
		b := &bytes.Buffer{} // TODO: use buffer pool here
		if err = body.Render(ctx, b, l); err != nil {
			return err
		}

		title, language, err := l.LocalizeWithTag(head.Title)
		if err != nil {
			return fmt.Errorf("cannot localize title: %w", err)
		}
		description, err := l.Localize(head.Description)
		if err != nil {
			return fmt.Errorf("cannot localize description: %w", err)
		}

		return pageTemplate.Execute(w, pageTemplateValues{
			Language:    language.String(),
			Title:       title,
			Description: description,
			Image:       head.Image,
			FavIconPNG:  head.FavIconPNG,
			Scripts:     head.Scripts,
			Stylesheets: head.Stylesheets,
			Main:        template.HTML(b.String()),
		})
	})
}

func PageStyleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	_, _ = io.Copy(w, bytes.NewReader(pageStyle))
}
