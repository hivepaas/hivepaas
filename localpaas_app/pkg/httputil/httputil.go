package httputil

import (
	"golang.org/x/text/language"

	"github.com/localpaas/localpaas/localpaas_app/pkg/translation"
)

// ParseRequestLang parse accept-language from http request.
// Return the best fit language for the input, if nothing matches the input, return the default language.
// E.g. Accept-Language: *
// E.g. Accept-Language: fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5
func ParseRequestLang(acceptLang string) translation.Lang {
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil {
		return translation.GetDefaultLang()
	}
	for _, tag := range tags {
		lang := translation.Lang(tag.String())
		if ok := translation.IsAvailable(lang); ok {
			return lang
		}

		base, _ := tag.Base()
		baseLang := translation.Lang(base.String())
		if ok := translation.IsAvailable(baseLang); ok {
			return baseLang
		}
	}
	return translation.GetDefaultLang()
}
