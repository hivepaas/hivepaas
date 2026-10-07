package sslproviderdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func pathsOf(errs hperrors.ValidationErrors) []string {
	var paths []string
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return paths
}

// A provider takes a name and a kind: an empty body was stored as a nameless
// setting of no kind.
func TestAnSSLProviderWithoutANameOrKindIsRefused(t *testing.T) {
	create := NewCreateSSLProviderReq()
	create.SSLProviderBaseReq = &SSLProviderBaseReq{}
	assert.ElementsMatch(t, []string{"name", "kind"}, pathsOf(create.Validate()))

	update := NewUpdateSSLProviderReq()
	update.SSLProviderBaseReq = &SSLProviderBaseReq{}
	assert.Subset(t, pathsOf(update.Validate()), []string{"name", "kind"})
}

func TestALetsEncryptProviderIsAccepted(t *testing.T) {
	req := NewCreateSSLProviderReq()
	req.SSLProviderBaseReq = &SSLProviderBaseReq{Name: "le", Kind: base.SSLProviderLetsEncrypt,
		Email: "ops@example.com", LetsEncrypt: &SSLProviderLetsEncryptReq{}}
	assert.Empty(t, pathsOf(req.Validate()))
}
