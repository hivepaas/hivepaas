package httpclient

import "net/http"

var (
	DefaultClient = func() *http.Client {
		return &http.Client{Transport: http.DefaultTransport}
	}()
)

// TODO: supports socks5 proxy
