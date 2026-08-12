package appsettingsuc

import (
	"context"
	"strings"

	dockerfile "github.com/hivepaas/dockerfile-generator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

const (
	templateGuideComment = "# NOTE: Please replace parameters enclosed in {{.Param}} " +
		"in this template before deployment.\n\n"
)

func (uc *UC) GetDockerfileTemplate(
	_ context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetDockerfileTemplateReq,
) (*appsettingsdto.GetDockerfileTemplateResp, error) {
	mainType, subkind, _ := strings.Cut(req.Type, "/")
	if rt, ok := dockerfile.ParseRuntime(mainType); ok {
		mainType = string(rt)
	}
	template := dockerfile.GetTemplate(dockerfile.Runtime(mainType), subkind)
	if template != "" {
		template = templateGuideComment + template
	}

	return &appsettingsdto.GetDockerfileTemplateResp{
		Data: &appsettingsdto.DockerfileTemplateResp{
			Template: template,
		},
	}, nil
}
