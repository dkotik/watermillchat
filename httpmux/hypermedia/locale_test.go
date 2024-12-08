package hypermedia_test

import (
	"testing"

	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"golang.org/x/text/language"
)

func TestSortLanguageTags(t *testing.T) {
	tags := []language.Tag{
		language.English, language.Dutch, language.German,
		language.Ukrainian,
	}
	hypermedia.SortLanguageTags(tags, "de, uk")

	if tags[0] != language.German {
		t.Error("first language is not German", tags)
	}

	if tags[1] != language.Ukrainian {
		t.Error("second language is not Ukrainian", tags)
	}
}
