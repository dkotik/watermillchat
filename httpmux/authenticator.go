package httpmux

const (
	unsafeNameCookieName = "watermillchatNaiveUnsafeName"
	unsafeIDCookieName   = "watermillchatNaiveUnsafeID"
)

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
