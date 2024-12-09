/*
Package httpmux injects front-end routing paths into a
[http.ServeMux] for rendering [watermillchat.Chat] view.
*/
package httpmux

import (
	"cmp"
	"context"
	_ "embed" // for media files and templates
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"github.com/dkotik/watermillchat/ollama"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/exp/rand"
)

//go:embed page.css
var stylesheet []byte

//go:embed hypermedia/script/post.js
var javascriptPost []byte

type Middleware func(http.Handler) http.Handler

func New(withOptions ...Option) (mux *http.ServeMux, err error) {
	o := &options{
		PageHead: hypermedia.Head{
			Title: &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "watermillchat.page.title",
					Other: "Watermill Chat",
				},
			},
			Description: &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "watermillchat.page.description",
					Other: "Watermill chat description.", // TODO: fill out
				},
			},
		},
		// TODO: make authenticator configurable
		Authenticator: NaiveBearerHeaderAuthenticatorUnsafe,
		Localization:  i18n.NewBundle(hypermedia.DefaultLanguage),
	}
	for _, option := range withOptions {
		if err = option.initializeMux(o); err != nil {
			return nil, fmt.Errorf("unable to mount Datastar front-end to HTTP multiplexer: %w", err)
		}
	}

	chat, err := watermillchat.NewChat(o.ChatOptions...)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(o.Prefix, "/") {
		o.Prefix = "/" + o.Prefix
	}
	if !strings.HasSuffix(o.Prefix, "/") {
		o.Prefix = o.Prefix + "/"
	}

	mux = cmp.Or(o.Base, &http.ServeMux{})
	o.PageHead.FavIconPNG = hypermedia.AddFavIconIfAbsent(mux, nil)
	o.PageHead.Scripts = append(o.PageHead.Scripts, o.Prefix+"post.js")
	o.PageHead.Scripts = append(o.PageHead.Scripts, o.Prefix+"datastar.js")
	mux.HandleFunc(o.Prefix+"datastar.js", hypermedia.DatastarHandler)
	mux.HandleFunc(o.Prefix+"datastar.js.map", hypermedia.DatastarMapHandler)
	mux.Handle(o.Prefix+"post.js", hypermedia.NewAsset("text/javascript", javascriptPost))
	o.PageHead.StyleSheets = append(o.PageHead.StyleSheets, o.Prefix+"style.css")
	mux.Handle(o.Prefix+"style.css", hypermedia.NewAsset("text/css", stylesheet))

	page := hypermedia.NewPageRenderer(o.PageHead)
	notFound := hypermedia.NewStaticPage(o.RenderingContext, page(hypermedia.ErrNotFound), o.Localization)

	errorHandler := hypermedia.ErrorHandlerWithLogger(
		hypermedia.NewErrorPageHandler(
			o.Localization,
			o.PageHead,
			[]hypermedia.RenderableError{
				hypermedia.ErrNotFound,
				hypermedia.ErrInternalServerError,
			}), o.Logger)

	mux.HandleFunc(o.Prefix+"{roomName}/messages", NewRoomMessagesHandler(
		chat,
		NewRoomSelectorFromURL("roomName"),
		errorHandler,
	))
	mux.Handle(o.Prefix+"send", o.Authenticator(NewMessageSendHandler(
		chat,
		hypermedia.ErrorHandlerWithLogger(hypermedia.PlainTextErrorHandler, o.Logger),
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
			http.Redirect(w, r, o.Prefix+string(result), http.StatusTemporaryRedirect)
		},
	)
	mux.HandleFunc(o.Prefix+"index.html", randomRoomRedirectSelector)
	mux.HandleFunc(o.Prefix+"{$}", randomRoomRedirectSelector)

	mux.HandleFunc(o.Prefix+"{roomName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: replace with RoomSelector
		roomName := strings.TrimSpace(r.PathValue("roomName"))
		if roomName == "" {
			panic("not found") // TODO: replace with hypermedia.ErrNotFound
			// return
		}
		hypermedia.NewPage(page(RoomRenderer{
			RoomName:        roomName,
			MessageSendPath: o.Prefix + "send",
		}), errorHandler, o.Localization).ServeHTTP(w, r)
	}))

	// show a 404 page for everything else
	mux.Handle(o.Prefix, notFound)

	bot := ollama.New("", "")
	go bot.JoinChat(context.Background(), chat, "Ollama", "ollama")

	return mux, nil
}
