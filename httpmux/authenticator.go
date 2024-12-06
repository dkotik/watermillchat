package httpmux

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/dkotik/watermillchat"
)

func NewNaiveBearerHeaderAuthenticatorUnsafe(next http.Handler) http.Handler {
	slog.Warn("an HTTP service is running with naive unsafe header token authenticator for demonstration purposes; it must never be used in production")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id, name, ok := strings.Cut(r.Header.Get("Authorization"), ":")
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			id = strings.TrimSpace(strings.TrimPrefix(id, `Bearer `))
			if id == "" {
				next.ServeHTTP(w, r)
				return
			}
			name = strings.TrimSpace(name)
			if name == "" {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(
				watermillchat.ContextWithIdentity(
					r.Context(),
					watermillchat.Identity{
						ID:   id,
						Name: name,
					},
				),
			))
		})
}

// type Authenticator interface {
// 	EstablishIdentity(*http.Request) (watermillchat.Identity, error)
// }

// type AuthenticatorFunc func(*http.Request) (watermillchat.Identity, error)

// func (f AuthenticatorFunc) EstablishIdentity(r *http.Request) (watermillchat.Identity, error) {
// 	return f(r)
// }

// func NewAuthenticationMiddleware(
// 	authenticator Authenticator,
// 	unauthenticatedHandler http.Handler,
// 	logger *slog.Logger,
// ) Middleware {
// 	if authenticator == nil {
// 		panic("cannot use a <nil> authenticator")
// 	}
// 	if unauthenticatedHandler == nil {
// 		panic("cannot use a <nil> unauthenticated handler")
// 	}
// 	if logger == nil {
// 		logger = slog.Default()
// 	}
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			identity, err := authenticator.EstablishIdentity(r)
// 			if err != nil {
// 				unauthenticatedHandler.ServeHTTP(w, r)
// 				if !errors.Is(err, ErrUnauthenticatedRequest) {
// 					logger.Error("authentication error", slog.Any("error", err))
// 				}
// 				return
// 			}
// 			next.ServeHTTP(w, r.WithContext(watermillchat.ContextWithIdentity(r.Context(), identity)))
// 		})
// 	}
// }

// func NewNaiveCookieAuthenticatorUnsafe(view http.Handler) http.HandlerFunc {
// 	extractor := func(r *http.Request) (id string, name string, err error) {
// 		cookie, err := r.Cookie(unsafeIDCookieName)
// 		if err != nil {
// 			return err
// 		}
// 		id = cookie.Value
// 		if id == "" {
// 			return "", "", http.ErrNoCookie
// 		}

// 		cookie, err = r.Cookie(unsafeNameCookieName)
// 		if err != nil {
// 			return "", "", http.ErrNoCookie
// 		}
// 		name = cookie.Value
// 		if name == "" {
// 			return http.ErrNoCookie
// 		}
// 		return id, name, nil
// 	}

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id, name, err := extractor(r)
// 		if err != nil {
// 			if !errors.Is(err, http.ErrNoCookie) {
// 				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 				slog.Error("cookie authenticator failed", slog.Any("error", err))
// 				return
// 			}
// 		}
// 	}
// }
