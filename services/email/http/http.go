package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/httpclient"
)

//nolint:gocognit
func SendMail(
	_ context.Context,
	conf *entity.EmailHTTP,
	recipients []string,
	subject string,
	content string,
) (err error) {
	password, err := conf.Password.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}

	var req *http.Request
	fromAddressField := gofn.Coalesce(conf.FieldMapping.FromAddress, "fromAddress")
	fromNameField := gofn.Coalesce(conf.FieldMapping.FromName, "fromName")
	toAddressField := gofn.Coalesce(conf.FieldMapping.ToAddress, "toAddress")
	toAddressesField := gofn.Coalesce(conf.FieldMapping.ToAddresses, "toAddresses")
	subjectField := gofn.Coalesce(conf.FieldMapping.Subject, "subject")
	contentField := gofn.Coalesce(conf.FieldMapping.Content, "content")
	passwordField := gofn.Coalesce(conf.FieldMapping.Password, "password")

	contentType := conf.ContentType
	switch conf.Method { //nolint:exhaustive
	case base.HTTPMethodPost, base.HTTPMethodPut:
		bodyMap := make(map[string]any)
		bodyMap[fromAddressField] = conf.Username
		if fromNameField != "" {
			bodyMap[fromNameField] = conf.DisplayName
		}
		bodyMap[subjectField] = subject
		bodyMap[contentField] = content
		if passwordField != "" {
			bodyMap[passwordField] = password
		}

		if contentType == "application/json" { //nolint:nestif
			if len(recipients) == 1 && toAddressesField != "" {
				bodyMap[toAddressField] = recipients[0]
			} else {
				bodyMap[toAddressesField] = recipients
			}

			dataBytes, err := json.Marshal(bodyMap)
			if err != nil {
				return apperrors.New(err)
			}

			req, err = http.NewRequest(string(conf.Method), conf.Endpoint, bytes.NewBuffer(dataBytes))
			if err != nil {
				return apperrors.New(err)
			}
		} else {
			contentType = gofn.Coalesce(contentType, "application/x-www-form-urlencoded")

			formValues := url.Values{}
			for k, v := range bodyMap {
				formValues.Add(k, fmt.Sprintf("%v", v))
			}
			if len(recipients) == 1 && toAddressesField != "" {
				formValues.Add(toAddressField, recipients[0])
			} else {
				formValues.Add(gofn.Coalesce(toAddressesField, toAddressField), strings.Join(recipients, ","))
			}

			req, err = http.NewRequest(string(conf.Method), conf.Endpoint, strings.NewReader(formValues.Encode()))
			if err != nil {
				return apperrors.New(err)
			}
		}

		req.Header.Set("Content-Type", contentType)

	case base.HTTPMethodGet:
		req, err = http.NewRequest(string(conf.Method), conf.Endpoint, nil)
		if err != nil {
			return apperrors.New(err)
		}

		q := req.URL.Query()
		q.Add(fromAddressField, conf.Username)
		if fromNameField != "" {
			q.Add(fromNameField, conf.DisplayName)
		}
		if len(recipients) == 1 && toAddressesField != "" {
			q.Add(toAddressField, recipients[0])
		} else {
			q.Add(gofn.Coalesce(toAddressesField, toAddressField), strings.Join(recipients, ","))
		}
		q.Add(subjectField, subject)
		q.Add(contentField, content)
		if passwordField != "" {
			q.Add(passwordField, password)
		}

		req.URL.RawQuery = q.Encode()

	default:
		return fmt.Errorf("%w: unsupported method: %s", apperrors.ErrUnsupported, conf.Method)
	}

	// TODO: support secrets within the header values?
	for k, v := range conf.Headers {
		req.Header.Set(k, v)
	}

	resp, err := httpclient.DefaultClient.Do(req)
	if err != nil {
		return apperrors.New(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: sending email via HTTP requests failed with status: %s",
			apperrors.ErrActionFailed, resp.Status)
	}
	return nil
}
