package acme

import (
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/providers/http/webroot"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func NewHTTP01Provider(webrootDir string) (provider challenge.Provider, err error) {
	http01Provider, err := webroot.NewHTTPProvider(webrootDir)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create http-01 provider for webroot")
	}
	return http01Provider, nil
}
