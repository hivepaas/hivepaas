package traefikserviceimpl

import (
	"context"
	"os"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) RemoveAppConfig(
	ctx context.Context,
	db database.IDB,
	req *traefikservice.RemoveAppConfigReq,
) (*traefikservice.RemoveAppConfigResp, error) {
	// Clean from Swarm Service
	if req.Service != nil && req.Service.Spec.Labels != nil {
		for k := range req.Service.Spec.Labels {
			if strings.HasPrefix(k, "traefik.") {
				delete(req.Service.Spec.Labels, k)
			}
		}
	}

	// Clean file
	err := os.Remove(req.App.TraefikConfigPath())
	if err != nil && !os.IsNotExist(err) {
		return nil, apperrors.Wrap(err)
	}

	return &traefikservice.RemoveAppConfigResp{}, nil
}
