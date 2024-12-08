package hypermedia

import (
	"slices"

	"golang.org/x/text/language"
)

var DefaultLanguage = language.AmericanEnglish

// SortLanguageTags prioritizes a list of language tags
// based on HTTP Accept-Language header definition as used by
// [language.ParseAcceptLanguage]. More desirable languages
// are moved to the front.
func SortLanguageTags(target []language.Tag, acceptLanguageHeader string) error {
	priority := make(map[language.Tag]float32)
	for _, initial := range target {
		priority[initial] = 0
	}

	tags, qs, err := language.ParseAcceptLanguage(acceptLanguageHeader)
	if err != nil {
		return err
	}
	for i, tag := range tags {
		priority[tag] = qs[i]
	}

	slices.SortFunc(target, func(a, b language.Tag) int {
		if priority[a] > priority[b] {
			return -1
		}
		if priority[a] >= priority[b] {
			return 1
		}
		return 0 // incomparable?
	})
	return nil
}
