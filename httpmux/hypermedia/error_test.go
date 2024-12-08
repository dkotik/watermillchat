package hypermedia_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func TestDefaultErrorHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = errors.New("test")

	hypermedia.DefaultErrorHandler.HandlerError(recorder, request, err)
	defer recorder.Result().Body.Close()

	// b := &bytes.Buffer{}
	// _, _ = io.Copy(b, recorder.Result().Body)
	// t.Fatal(b)
}

func TestErrorRendering(t *testing.T) {
	b := &bytes.Buffer{}
	bundle := i18n.NewBundle(hypermedia.DefaultLanguage)
	err := hypermedia.ErrNotFound.Render(context.Background(), b,
		i18n.NewLocalizer(bundle, hypermedia.DefaultLanguage.String()))
	if err != nil {
		t.Fatal(err)
	}
	// t.Fatal(b.String())
}
