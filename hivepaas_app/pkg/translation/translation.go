package translation

import (
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Lang string

const (
	LangEn Lang = "en"

	defaultLang = LangEn
)

const (
	// emptyMsg special case: use `-` to mark an empty message
	// as go-i18n doesn't support an empty one.
	emptyMsg = "-"
)

var (
	// localeMap - mapping from lang to go-i18n lang
	localeMap = map[Lang]language.Tag{
		LangEn: language.English,
	}

	//go:embed messages/*
	messages embed.FS

	// localizerMap - mapping from lang to go-i18n localizer
	localizerMap = func() map[Lang]*i18n.Localizer {
		localizerEn, err := load("messages", LangEn)
		if err != nil {
			panic(err)
		}
		return map[Lang]*i18n.Localizer{LangEn: localizerEn}
	}()
)

func (l Lang) String() string {
	return string(l)
}

// GetDefaultLang return the default language
func GetDefaultLang() Lang {
	return defaultLang
}

func IsAvailable(lang Lang) bool {
	_, ok := localizerMap[lang]
	return ok
}

// LocalizeEx gets translated message for the given ID
func LocalizeEx(lang Lang, msgID string, templateData map[string]any) (string, error) {
	localizer, ok := localizerMap[lang]
	if !ok {
		localizer = localizerMap[defaultLang]
	}

	cnf := &i18n.LocalizeConfig{
		MessageID: msgID,
	}
	if templateData != nil {
		cnf.TemplateData = templateData
	}

	msg, err := localizer.Localize(cnf)
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	if msg == emptyMsg {
		return "", nil
	}
	return msg, nil
}

// Localize gets translated message for the given ID
func Localize(lang Lang, msgID string) (string, error) {
	return LocalizeEx(lang, msgID, nil)
}

// LocalizeOrKey returns translated message or the input key if missing translation
func LocalizeOrKey(lang Lang, msgID string, templateData map[string]any) string {
	msg, _ := LocalizeEx(lang, msgID, templateData)
	if msg == "" {
		return msgID
	}
	return msg
}
