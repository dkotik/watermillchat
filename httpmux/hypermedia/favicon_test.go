package hypermedia_test

import (
	"net/http"
	"testing"

	"github.com/dkotik/watermillchat/httpmux/hypermedia"
)

func TestFavIcon(t *testing.T) {
	mux := &http.ServeMux{}
	mux.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("should not hit the root node when matching")
		},
	))
	p := hypermedia.AddFavIcon(mux, nil)
	if p != "/favicon.png" {
		t.Fatal("unexpected icon path:", p)
	}

	p = hypermedia.AddFavIconIfAbsent(mux, nil)
	if p != "/favicon.png" {
		t.Fatal("unexpected icon path:", p)
	}

	r, err := http.NewRequest(http.MethodGet, p, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, p = mux.Handler(r)
	if p != "/favicon.png" {
		t.Fatal("unexpected icon path:", p)
	}
}
