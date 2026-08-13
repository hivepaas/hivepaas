package imagebuildagentdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

type ImageBuildReq struct {
	TaskID  string
	AppID   string
	SendLog func(frames []*tasklog.LogFrame) error

	imagebuildservice.ImageBuildReq
}

type ImageBuildResp struct {
	imagebuildservice.ImageBuildResp
}
