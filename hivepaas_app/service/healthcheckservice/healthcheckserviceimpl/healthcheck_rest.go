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

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httpclient"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jsonutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
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
	periodicJob := data.PeriodicSetting.MustAsPeriodicJob()
	healthchk := data.Healthcheck.REST
	if data.Output.REST == nil {
		data.Output.REST = &entity.TaskPeriodicHealthcheckOutputREST{}
	}

	reqCtx := ctx
	if periodicJob.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, periodicJob.Timeout.ToDuration())
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
		return hperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", gofn.Coalesce(healthchk.ContentType, restContentTypeDefault))

	resp, err := httpclient.DefaultClient.Do(req)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	data.Output.REST.ReturnCode = resp.StatusCode
	if len(healthchk.ReturnCode) > 0 && !gofn.Contain(healthchk.ReturnCode, resp.StatusCode) {
		return hperrors.Wrap(hperrors.ErrActionFailed)
	}

	if healthchk.ReturnText != nil || healthchk.ReturnJSON != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return hperrors.Wrap(err)
		}
		bodyStr := reflectutil.UnsafeBytesToStr(body)
		data.Output.REST.ReturnText = strutil.CutShort(bodyStr, restBodySavingMaxLen, "...")

		if healthchk.ReturnText != nil {
			if healthchk.ReturnText.Exact != "" && healthchk.ReturnText.Exact != bodyStr {
				return hperrors.Wrap(hperrors.ErrActionFailed)
			}
			if healthchk.ReturnText.Regex != "" {
				matched, _ := regexp.MatchString(healthchk.ReturnText.Regex, bodyStr)
				if !matched {
					return hperrors.Wrap(hperrors.ErrActionFailed)
				}
			}
		}
		if healthchk.ReturnJSON != nil {
			var actualObj any
			err := json.Unmarshal(body, &actualObj)
			if err != nil {
				return hperrors.Wrap(hperrors.ErrActionFailed)
			}

			if healthchk.ReturnJSON.Exact != nil {
				if !reflect.DeepEqual(actualObj, healthchk.ReturnJSON.Exact) {
					return hperrors.Wrap(hperrors.ErrActionFailed)
				}
			}
			if healthchk.ReturnJSON.Contain != nil {
				if !jsonutil.Contains(actualObj, healthchk.ReturnJSON.Contain) {
					return hperrors.Wrap(hperrors.ErrActionFailed)
				}
			}
		}
	}

	return nil
}
