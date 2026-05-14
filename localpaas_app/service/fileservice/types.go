package fileservice

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type GetDownloadURLReq struct {
	File         *entity.Setting
	RequireLogin bool
	Expiration   time.Duration
	CloudPresign bool
	URLPath      string // if empty, `download` is used
	ViewInline   bool
}

type GetDownloadURLResp struct {
	URL string
}
