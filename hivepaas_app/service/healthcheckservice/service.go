package healthcheckservice

import (
	"context"
)

type Service interface {
	Healthcheck(ctx context.Context, req *HealthcheckReq) (*HealthcheckResp, error)
}
