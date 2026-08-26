package cfutil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httputil"
)

func VerifyToken(
	ctx context.Context,
	token string,
) error {
	data, err := httputil.HTTPGet(ctx, "https://api.cloudflare.com/client/v4/user/tokens/verify",
		func(httpReq *http.Request) {
			httpReq.Header.Set("Authorization", "Bearer "+token)
			httpReq.Header.Set("Content-Type", "application/json")
		})
	if err != nil {
		return apperrors.Wrap(apperrors.ErrTokenInvalid).WithCause(err)
	}

	var cloudflareResp struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(data, &cloudflareResp); err != nil {
		return apperrors.Wrap(err)
	}

	if !cloudflareResp.Success {
		return apperrors.Wrap(apperrors.ErrTokenInvalid).WithMsgLog(
			"Cloudflare token verification response success was false")
	}

	return nil
}
