package registrydto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

// ProbeDomainReq asks what answers at an address. The domain is in the body
// rather than read from the stored setting, so that the dashboard can check one
// the operator has typed but not saved.
type ProbeDomainReq struct {
	Domain string `json:"domain"`
}

type ProbeDomainResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *DomainProbeResp `json:"data"`
}

type DomainProbeResp struct {
	Reached  bool     `json:"reached"`
	Proxied  bool     `json:"proxied"`
	Evidence []string `json:"evidence"`
}

// PushCheckReq asks for an upload of Bytes through the public domain. Zero means
// the default, which is above the limit a proxy on a free plan imposes.
type PushCheckReq struct {
	Bytes int64 `json:"bytes"`
}

type PushCheckResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *PushCheckDataRes `json:"data"`
}

type PushCheckDataRes struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Detail     string `json:"detail"`
	ElapsedMs  int64  `json:"elapsedMs"`
}

type RotateCredentialResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *RotateCredentialDataRes `json:"data"`
}

type RotateCredentialDataRes struct {
	// GraceEndsAt is when the previous password stops working, which is the date
	// every app has to be redeployed by.
	GraceEndsAt time.Time `json:"graceEndsAt"`
}
