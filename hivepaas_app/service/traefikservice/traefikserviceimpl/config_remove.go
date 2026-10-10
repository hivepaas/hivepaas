package traefikserviceimpl

import (
	"context"
	"os"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

	// Clean file, and the one named after the app's key, from before
	for _, path := range []string{req.App.TraefikConfigPath(), req.App.LegacyTraefikConfigPath()} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, hperrors.Wrap(err)
		}
	}

	return &traefikservice.RemoveAppConfigResp{}, nil
}
