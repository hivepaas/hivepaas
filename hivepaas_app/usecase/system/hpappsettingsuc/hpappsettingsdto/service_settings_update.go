package hpappsettingsdto

import (
	"strings"
	"time"
	"unicode"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	mainServiceReplicasMin = 1
	mainServiceReplicasMax = 100

	workerServiceReplicasMin    = 0
	workerServiceReplicasMax    = 100
	workerServiceConcurrencyMin = 1
	workerServiceConcurrencyMax = 100

	taskCheckIntervalMin  = timeutil.Duration(30 * time.Second)
	taskCheckIntervalMax  = timeutil.Duration(24 * time.Hour)
	taskCreateIntervalMin = timeutil.Duration(30 * time.Second)
	taskCreateIntervalMax = timeutil.Duration(24 * time.Hour)

	periodicBaseIntervalMin = timeutil.Duration(1 * time.Second)
	periodicBaseIntervalMax = timeutil.Duration(24 * time.Hour)
	periodicBatchSizeMin    = 1
	periodicBatchSizeMax    = 10000
)

type UpdateServiceSettingsReq struct {
	*ServiceSettingsBaseReq
	UpdateVer int `json:"updateVer"`
}

type ServiceSettingsBaseReq struct {
	AppSettings      HivePaaSAppSettingsReq      `json:"appSettings"`
	WorkerSettings   HivePaaSWorkerSettingsReq   `json:"workerSettings"`
	TaskSettings     HivePaaSTaskSettingsReq     `json:"taskSettings"`
	PeriodicSettings HivePaaSPeriodicSettingsReq `json:"periodicSettings"`
	ProxySettings    HivePaaSProxySettingsReq    `json:"proxySettings"`
}

func (req *ServiceSettingsBaseReq) ToEntity() *entity.HivePaaSService {
	return &entity.HivePaaSService{
		AppSettings:      *req.AppSettings.ToEntity(),
		WorkerSettings:   *req.WorkerSettings.ToEntity(),
		TaskSettings:     *req.TaskSettings.ToEntity(),
		PeriodicSettings: *req.PeriodicSettings.ToEntity(),
		ProxySettings:    *req.ProxySettings.ToEntity(),
	}
}

func (req *ServiceSettingsBaseReq) modifyRequest() (err error) {
	if err = req.ProxySettings.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (req *ServiceSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.AppSettings.validate(field+"appSettings")...)
	res = append(res, req.WorkerSettings.validate(field+"workerSettings")...)
	res = append(res, req.TaskSettings.validate(field+"taskSettings")...)
	res = append(res, req.PeriodicSettings.validate(field+"periodicSettings")...)
	res = append(res, req.ProxySettings.validate(field+"proxySettings")...)
	return res
}

type HivePaaSAppSettingsReq struct {
	Replicas int `json:"replicas"`
}

func (req *HivePaaSAppSettingsReq) ToEntity() *entity.HivePaaSAppSettings {
	return &entity.HivePaaSAppSettings{
		Replicas: req.Replicas,
	}
}

func (req *HivePaaSAppSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Replicas, true, mainServiceReplicasMin,
		mainServiceReplicasMax, field+"replicas")...)
	return res
}

type HivePaaSWorkerSettingsReq struct {
	Replicas           int  `json:"replicas"`
	Concurrency        int  `json:"concurrency"`
	RunWorkerInMainApp bool `json:"runWorkerInMainApp"`
}

func (req *HivePaaSWorkerSettingsReq) ToEntity() *entity.HivePaaSWorkerSettings {
	return &entity.HivePaaSWorkerSettings{
		Replicas:           req.Replicas,
		Concurrency:        req.Concurrency,
		RunWorkerInMainApp: req.RunWorkerInMainApp,
	}
}

func (req *HivePaaSWorkerSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateNumber(&req.Replicas, true, workerServiceReplicasMin,
		workerServiceReplicasMax, field+"replicas")...)
	res = append(res, basedto.ValidateNumber(&req.Concurrency, true, workerServiceConcurrencyMin,
		workerServiceConcurrencyMax, field+"concurrency")...)
	if req.Replicas == 0 && !req.RunWorkerInMainApp {
		res = append(res, vld.Must(false).OnError(
			vld.SetField(field+"runWorkerInMainApp", nil),
			vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
		))
	}
	return res
}

type HivePaaSTaskSettingsReq struct {
	TaskCheckInterval  timeutil.Duration `json:"taskCheckInterval"`
	TaskCreateInterval timeutil.Duration `json:"taskCreateInterval"`
}

func (req *HivePaaSTaskSettingsReq) ToEntity() *entity.HivePaaSTaskSettings {
	return &entity.HivePaaSTaskSettings{
		TaskCheckInterval:  req.TaskCheckInterval,
		TaskCreateInterval: req.TaskCreateInterval,
	}
}

func (req *HivePaaSTaskSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDuration(&req.TaskCheckInterval, true, taskCheckIntervalMin,
		taskCheckIntervalMax, field+"taskCheckInterval")...)
	res = append(res, basedto.ValidateDuration(&req.TaskCreateInterval, true, taskCreateIntervalMin,
		taskCreateIntervalMax, field+"taskCreateInterval")...)
	return res
}

type HivePaaSPeriodicSettingsReq struct {
	BaseInterval timeutil.Duration `json:"baseInterval"`
	BatchSize    int               `json:"batchSize"`
}

func (req *HivePaaSPeriodicSettingsReq) ToEntity() *entity.HivePaaSPeriodicSettings {
	return &entity.HivePaaSPeriodicSettings{
		BaseInterval: req.BaseInterval,
		BatchSize:    req.BatchSize,
	}
}

func (req *HivePaaSPeriodicSettingsReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDuration(&req.BaseInterval, true, periodicBaseIntervalMin,
		periodicBaseIntervalMax, field+"baseInterval")...)
	res = append(res, basedto.ValidateNumber(&req.BatchSize, true, periodicBatchSizeMin,
		periodicBatchSizeMax, field+"batchSize")...)
	return res
}

const (
	proxyProviderMaxLen = 50

	// proxyHopsMin and proxyHopsMax bound the X-Forwarded-For depth. One is the
	// smallest position that can hold a caller; the upper bound is not a real
	// topology so much as a guard - past a handful of hops the value is a typo,
	// and a depth longer than the chain makes every caller resolve to nothing and
	// share a single rate-limit bucket.
	proxyHopsMin = 1
	proxyHopsMax = 10
)

type HivePaaSProxySettingsReq struct {
	ProxyProvider string   `json:"proxyProvider"`
	TrustedIPs    []string `json:"trustedIPs"`
	ProxyHops     int      `json:"proxyHops"`
}

func (req *HivePaaSProxySettingsReq) ToEntity() *entity.HivePaaSProxySettings {
	return &entity.HivePaaSProxySettings{
		ProxyProvider: req.ProxyProvider,
		TrustedIPs:    req.TrustedIPs,
		ProxyHops:     req.ProxyHops,
	}
}

//nolint:unparam
func (req *HivePaaSProxySettingsReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.ProxyProvider = strings.ToLower(strings.TrimSpace(req.ProxyProvider))
	req.TrustedIPs = strings.FieldsFunc(strings.Join(req.TrustedIPs, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	// No proxy means the other two describe nothing, so they are dropped rather
	// than rejected: a form that still holds yesterday's values must not block the
	// operator from saying there is no proxy any more. Dropping them is also what
	// keeps the three fields from disagreeing - see validate.
	if req.ProxyProvider == "" {
		req.TrustedIPs = nil
		req.ProxyHops = 0
	}
	return nil
}

// validate keeps the three proxy fields consistent with each other.
//
// Declaring a proxy without saying which addresses it comes from is the shape
// that hurts: Traefik is then never told to trust anything, so every forwarded
// header is ignored, every caller behind the proxy is counted as one, and the
// rate limits meant to protect the install throttle it instead. The operator gets
// no signal at all - the settings look configured. Requiring the fields together
// is what turns that silent state into a form error.
func (req *HivePaaSProxySettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return res
	}
	if field != "" {
		field += "."
	}
	if req.ProxyProvider == "" {
		return res
	}

	res = append(res, basedto.ValidateStr(&req.ProxyProvider, true, 1,
		proxyProviderMaxLen, field+"proxyProvider")...)
	res = append(res, basedto.ValidateIPOrCIDRSlice(req.TrustedIPs, 1, field+"trustedIPs")...)
	res = append(res, basedto.ValidateNumber(&req.ProxyHops, true,
		proxyHopsMin, proxyHopsMax, field+"proxyHops")...)
	return res
}

func NewUpdateServiceSettingsReq() *UpdateServiceSettingsReq {
	return &UpdateServiceSettingsReq{}
}

func (req *UpdateServiceSettingsReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateServiceSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateServiceSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
