/*
Package httpmux injects front-end routing paths into a
[http.ServeMux] for rendering [watermillchat.Chat] view.
*/
package httpmux

import (
	"context"
	_ "embed" // for media files and templates
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/exp/rand"
)

//go:embed page.css
var stylesheet []byte

//go:embed hypermedia/script/post.js
var javascriptPost []byte

type Middleware func(http.Handler) http.Handler

type RenderingConfiguration struct {
	Context      context.Context
	PageHead     hypermedia.Head
	Localization *i18n.Bundle
}

type Configuration struct {
	Chat          *watermillchat.Chat
	HostName      string
	BaseMux       *http.ServeMux
	Prefix        string
	Authenticator Middleware
	Rendering     RenderingConfiguration
	Logger        *slog.Logger
}

func (c Configuration) Validate() (err error) {
	if c.Chat == nil {
		err = errors.Join(err, errors.New("missing Chat"))
	}
	if c.BaseMux == nil {
		err = errors.Join(err, errors.New("missing Base"))
	}
	if c.Authenticator == nil {
		err = errors.Join(err, errors.New("missing Authenticator"))
	}
	if c.Rendering.Context == nil {
		err = errors.Join(err, errors.New("missing rendering context"))
	}
	if c.Rendering.Localization == nil {
		err = errors.Join(err, errors.New("missing rendering localization"))
	}
	if c.Rendering.PageHead.Title == nil {
		err = errors.Join(err, errors.New("missing page head title"))
	}
	if c.Rendering.PageHead.Description == nil {
		err = errors.Join(err, errors.New("missing page head description"))
	}
	if c.Logger == nil {
		err = errors.Join(err, errors.New("missing Logger"))
	}
	return err
}

func New(c Configuration) (mux *http.ServeMux, err error) {
	if c.Chat == nil {
		c.Chat, err = watermillchat.NewChat()
		if err != nil {
			return nil, err
		}
	}
	if c.BaseMux == nil {
		c.BaseMux = &http.ServeMux{}
	}
	if !strings.HasPrefix(c.Prefix, "/") {
		c.Prefix = "/" + c.Prefix
	}
	if !strings.HasSuffix(c.Prefix, "/") {
		c.Prefix = c.Prefix + "/"
	}

	if c.Rendering.Context == nil {
		c.Rendering.Context = context.Background()
	}
	if c.Rendering.PageHead.Title == nil {
		c.Rendering.PageHead.Title = &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "watermillchat.page.title",
				Other: "Watermill Chat",
			},
		}
	}
	if c.Rendering.PageHead.Description == nil {
		c.Rendering.PageHead.Description = &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "watermillchat.page.description",
				Other: "Watermill chat description.", // TODO: fill out
			},
		}
	}
	if c.Rendering.Localization == nil {
		c.Rendering.Localization = i18n.NewBundle(hypermedia.DefaultLanguage)
	}

	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	if err = c.Validate(); err != nil {
		return nil, err
	}

	mux = c.BaseMux
	c.Rendering.PageHead.FavIconPNG = hypermedia.AddFavIconIfAbsent(mux, nil)
	c.Rendering.PageHead.Scripts = append(c.Rendering.PageHead.Scripts, c.Prefix+"post.js")
	c.Rendering.PageHead.Scripts = append(c.Rendering.PageHead.Scripts, c.Prefix+"datastar.js")
	mux.HandleFunc(c.Prefix+"datastar.js", hypermedia.DatastarHandler)
	mux.HandleFunc(c.Prefix+"datastar.js.map", hypermedia.DatastarMapHandler)
	mux.Handle(c.Prefix+"post.js", hypermedia.NewAsset("text/javascript", javascriptPost))
	c.Rendering.PageHead.StyleSheets = append(c.Rendering.PageHead.StyleSheets, c.Prefix+"style.css")
	mux.Handle(c.Prefix+"style.css", hypermedia.NewAsset("text/css", stylesheet))

	page := hypermedia.NewPageRenderer(c.Rendering.PageHead)
	notFound := hypermedia.NewStaticPage(c.Rendering.Context, page(hypermedia.ErrNotFound), c.Rendering.Localization)

	errorHandler := hypermedia.ErrorHandlerWithLogger(
		hypermedia.NewErrorPageHandler(
			c.Rendering.Localization,
			c.Rendering.PageHead,
			[]hypermedia.RenderableError{
				hypermedia.ErrNotFound,
				hypermedia.ErrInternalServerError,
			}), c.Logger)

	mux.HandleFunc(c.Prefix+"{roomName}/messages", NewRoomMessagesHandler(
		c.Chat,
		NewRoomSelectorFromURL("roomName"),
		errorHandler,
	))
	mux.Handle(c.Prefix+"send", c.Authenticator(NewMessageSendHandler(
		c.Chat,
		hypermedia.ErrorHandlerWithLogger(hypermedia.PlainTextErrorHandler, c.Logger),
	)))

	randomRoomRedirectSelector := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: extract room history
			// to avoid redirecting to an active room

			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			seed := rand.NewSource(uint64(time.Now().UnixNano()))
			random := rand.New(seed)

			result := make([]byte, 32)
			for i := range result {
				result[i] = charset[random.Intn(len(charset))]
			}
			http.Redirect(w, r, c.Prefix+string(result), http.StatusTemporaryRedirect)
		},
	)
	mux.HandleFunc(c.Prefix+"index.html", randomRoomRedirectSelector)
	mux.HandleFunc(c.Prefix+"{$}", randomRoomRedirectSelector)

	mux.HandleFunc(c.Prefix+"{roomName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: replace with RoomSelector
		roomName := strings.TrimSpace(r.PathValue("roomName"))
		if roomName == "" {
			panic("not found") // TODO: replace with hypermedia.ErrNotFound
			// return
		}
		hypermedia.NewPage(page(RoomRenderer{
			RoomName:        roomName,
			MessageSendPath: c.Prefix + "send",
		}), errorHandler, c.Rendering.Localization).ServeHTTP(w, r)
	}))

	// show a 404 page for everything else
	mux.Handle(c.Prefix, notFound)
	return mux, nil
}
