package hypermedia

import (
	_ "embed" // for default icon
	"fmt"
	"net/http"
)

//go:embed icon/favicon.png
var defaultFavIconPNG []byte

func AddFavIcon(mux *http.ServeMux, image []byte) string {
	if len(image) == 0 {
		mux.Handle("/favicon.png", NewAsset("image/png", defaultFavIconPNG))
		return "/favicon.png"
	}

	ct := http.DetectContentType(image)
	switch ct {
	case "image/png":
		mux.Handle("/favicon.png", NewAsset(ct, image))
		return "/favicon.png"
	default:
		panic(fmt.Errorf("icon extension not supported for content type: %s", ct))
	}
}

// TODO: requires test! does not seem to work
func AddFavIconIfAbsent(mux *http.ServeMux, image []byte) (found string) {
	r, err := http.NewRequest(http.MethodGet, "/favicon.png", nil)
	if err != nil {
		panic(err)
	}
	_, found = mux.Handler(r)
	if found != "" {
		return
	}

	r.URL.Path = "/favicon.svg"
	_, found = mux.Handler(r)
	if found != "" {
		return "/favicon.svg"
	}

	r.URL.Path = "/favicon.ico"
	_, found = mux.Handler(r)
	if found != "" {
		return "/favicon.ico"
	}

	return AddFavIcon(mux, image)
}
