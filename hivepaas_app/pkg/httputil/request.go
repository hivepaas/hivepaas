package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

var (
	ErrHTTPStatus = errors.New("http status error")
)

type RequestSetupFunc func(r *http.Request)

// HTTPGet sends a GET request to get data from a URL
func HTTPGet(ctx context.Context, url string, reqFuncs ...RequestSetupFunc) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	for _, reqFunc := range reqFuncs {
		reqFunc(req)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 { //nolint:mnd
		return nil, tracerr.Wrap(fmt.Errorf("%w: %s", ErrHTTPStatus, res.Status))
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return data, nil
}

// HTTPPost sends a POST request to get data from a URL
func HTTPPost(
	ctx context.Context,
	url string,
	body io.Reader,
	reqFuncs ...RequestSetupFunc,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	for _, reqFunc := range reqFuncs {
		reqFunc(req)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 { //nolint:mnd
		return nil, tracerr.Wrap(fmt.Errorf("%w: %s", ErrHTTPStatus, res.Status))
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return data, nil
}
