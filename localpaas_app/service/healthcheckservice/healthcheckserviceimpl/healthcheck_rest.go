package healthcheckserviceimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/httpclient"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jsonutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
)

const (
	restContentTypeDefault = "application/json"
	restBodySavingMaxLen   = 500
)

//nolint:gocognit
func (s *service) doHealthcheckREST(
	ctx context.Context,
	data *healthcheckData,
) (err error) {
	healthchk := data.Healthcheck.REST
	if data.Output.REST == nil {
		data.Output.REST = &entity.TaskHealthcheckOutputREST{}
	}

	reqCtx := ctx
	if data.Healthcheck.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, data.Healthcheck.Timeout.ToDuration())
		defer cancel()
		reqCtx = ctx
	}

	method := gofn.Coalesce(healthchk.Method, "GET")
	var input io.Reader
	if healthchk.Body != "" {
		input = strings.NewReader(healthchk.Body)
	}
	req, err := http.NewRequestWithContext(reqCtx, string(method), healthchk.URL, input)
	if err != nil {
		return apperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", gofn.Coalesce(healthchk.ContentType, restContentTypeDefault))

	resp, err := httpclient.DefaultClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	data.Output.REST.ReturnCode = resp.StatusCode
	if len(healthchk.ReturnCode) > 0 && !gofn.Contain(healthchk.ReturnCode, resp.StatusCode) {
		return apperrors.Wrap(apperrors.ErrActionFailed)
	}

	//nolint:nestif
	if healthchk.ReturnText != nil || healthchk.ReturnJSON != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return apperrors.Wrap(err)
		}
		bodyStr := reflectutil.UnsafeBytesToStr(body)
		data.Output.REST.ReturnText = strutil.CutShort(bodyStr, restBodySavingMaxLen, "...")

		if healthchk.ReturnText != nil {
			if healthchk.ReturnText.Exact != "" && healthchk.ReturnText.Exact != bodyStr {
				return apperrors.Wrap(apperrors.ErrActionFailed)
			}
			if healthchk.ReturnText.Regex != "" {
				matched, _ := regexp.MatchString(healthchk.ReturnText.Regex, bodyStr)
				if !matched {
					return apperrors.Wrap(apperrors.ErrActionFailed)
				}
			}
		}
		if healthchk.ReturnJSON != nil {
			var actualObj any
			err := json.Unmarshal(body, &actualObj)
			if err != nil {
				return apperrors.Wrap(apperrors.ErrActionFailed)
			}

			if healthchk.ReturnJSON.Exact != nil {
				if !reflect.DeepEqual(actualObj, healthchk.ReturnJSON.Exact) {
					return apperrors.Wrap(apperrors.ErrActionFailed)
				}
			}
			if healthchk.ReturnJSON.Contain != nil {
				if !jsonutil.Contains(actualObj, healthchk.ReturnJSON.Contain) {
					return apperrors.Wrap(apperrors.ErrActionFailed)
				}
			}
		}
	}

	return nil
}
