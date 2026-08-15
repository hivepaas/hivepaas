package hpappsettingsdto

import (
	"strings"
	"time"
	"unicode"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		return apperrors.Wrap(err)
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

type HivePaaSProxySettingsReq struct {
	ProxyProvider string   `json:"proxyProvider"`
	TrustedIPs    []string `json:"trustedIPs"`
}

func (req *HivePaaSProxySettingsReq) ToEntity() *entity.HivePaaSProxySettings {
	return &entity.HivePaaSProxySettings{
		ProxyProvider: req.ProxyProvider,
		TrustedIPs:    req.TrustedIPs,
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
	return nil
}

func (req *HivePaaSProxySettingsReq) validate(_ string) (res []vld.Validator) {
	return res
}

func NewUpdateServiceSettingsReq() *UpdateServiceSettingsReq {
	return &UpdateServiceSettingsReq{}
}

func (req *UpdateServiceSettingsReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateServiceSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateServiceSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
