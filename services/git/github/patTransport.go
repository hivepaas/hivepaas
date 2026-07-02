package github

import (
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

type PatTransport struct {
	tr    http.RoundTripper
	token string
}

func NewPatTransport(tr http.RoundTripper, token string) *PatTransport {
	return &PatTransport{
		tr:    tr,
		token: token,
	}
}

// RoundTrip implements http.RoundTripper interface.
func (t *PatTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "token "+t.token)
	// req.Header.Add("Accept", "application/vnd.github.v3+json")

	resp, err := t.tr.RoundTrip(req)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp, nil
}
