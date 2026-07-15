package appsettingsdto

import (
	"fmt"
	"strings"
	"unicode"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/traefik"
)

type UpdateAppHttpSettingsReq struct {
	ProjectID      string       `json:"-"`
	AppID          string       `json:"-"`
	ExposePublicly bool         `json:"exposePublicly"`
	Domains        []*DomainReq `json:"domains"`
	UpdateVer      int          `json:"updateVer"`
}

func (req *UpdateAppHttpSettingsReq) ToEntity() *entity.AppHttpSettings {
	return &entity.AppHttpSettings{
		ExposePublicly: req.ExposePublicly,
		Domains: gofn.MapSlice(req.Domains, func(r *DomainReq) *entity.AppDomain {
			return r.ToEntity()
		}),
	}
}

type DomainReq struct {
	Enabled              bool                         `json:"enabled"`
	Domain               string                       `json:"domain"`
	DomainRedirect       string                       `json:"domainRedirect"`
	SSLCert              basedto.ObjectIDReq          `json:"sslCert"`
	ContainerPort        int                          `json:"containerPort"`
	ForceHttps           bool                         `json:"forceHttps"`
	LBConfig             *HTTPLBConfigReq             `json:"lbConfig"`
	BasicAuth            *HTTPBasicAuthConfigReq      `json:"basicAuth"`
	ClientConfig         *HTTPClientConfigReq         `json:"clientConfig"`
	HeaderConfig         *HTTPHeaderConfigReq         `json:"headerConfig"`
	CompressionConfig    *HTTPCompressionConfigReq    `json:"compressionConfig"`
	RateLimitConfig      *HTTPRateLimitConfigReq      `json:"rateLimitConfig"`
	PathRewriteConfig    *HTTPPathRewriteConfigReq    `json:"pathRewriteConfig"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigReq `json:"circuitBreakerConfig"`
	Paths                []*HTTPPathConfigReq         `json:"paths"`
}

func (req *DomainReq) ToEntity() *entity.AppDomain {
	return &entity.AppDomain{
		Enabled:              req.Enabled,
		Domain:               req.Domain,
		DomainRedirect:       req.DomainRedirect,
		SSLCert:              entity.ObjectID{ID: req.SSLCert.ID},
		ContainerPort:        req.ContainerPort,
		ForceHttps:           req.ForceHttps,
		LBConfig:             req.LBConfig.ToEntity(),
		BasicAuth:            req.BasicAuth.ToEntity(),
		ClientConfig:         req.ClientConfig.ToEntity(),
		HeaderConfig:         req.HeaderConfig.ToEntity(),
		CompressionConfig:    req.CompressionConfig.ToEntity(),
		RateLimitConfig:      req.RateLimitConfig.ToEntity(),
		PathRewriteConfig:    req.PathRewriteConfig.ToEntity(),
		CircuitBreakerConfig: req.CircuitBreakerConfig.ToEntity(),
		Paths: gofn.MapSlice(req.Paths, func(item *HTTPPathConfigReq) *entity.HTTPPathConfig {
			return item.ToEntity()
		}),
	}
}

//nolint:unparam
func (req *DomainReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	if err := req.LBConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.BasicAuth.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.ClientConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.HeaderConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.CompressionConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.RateLimitConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.PathRewriteConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := req.CircuitBreakerConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	for _, pathReq := range req.Paths {
		if err := pathReq.modifyRequest(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

//nolint:unparam
func (req *DomainReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateDomain(&req.Domain, true, base.DomainNameMaxLen,
		false, field+"domain")...)
	res = append(res, basedto.ValidateDomain(&req.DomainRedirect, false, base.DomainNameMaxLen,
		false, field+"domainRedirect")...)
	res = append(res, basedto.ValidatePort(&req.ContainerPort, true, 1, field+"containerPort")...)

	res = append(res, req.LBConfig.validate(field+"lbConfig")...)
	res = append(res, req.BasicAuth.validate(field+"basicAuth")...)
	res = append(res, req.ClientConfig.validate(field+"clientConfig")...)
	res = append(res, req.HeaderConfig.validate(field+"headerConfig")...)
	res = append(res, req.CompressionConfig.validate(field+"compressionConfig")...)
	res = append(res, req.RateLimitConfig.validate(field+"rateLimitConfig")...)
	res = append(res, req.PathRewriteConfig.validate(field+"pathRewriteConfig")...)
	res = append(res, req.CircuitBreakerConfig.validate(field+"circuitBreakerConfig")...)
	for i, pathReq := range req.Paths {
		res = append(res, pathReq.validate(field+fmt.Sprintf("paths[%v]", i))...)
	}
	return res
}

type HTTPLBConfigReq struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

func (r *HTTPLBConfigReq) ToEntity() *entity.HTTPLBConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPLBConfig{
		Strategy: r.Strategy,
	}
}

//nolint:unparam
func (r *HTTPLBConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (r *HTTPLBConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&r.Strategy, false, traefik.AllLBStrategies, field+"strategy")...)
	return res
}

type HTTPBasicAuthConfigReq struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

func (r *HTTPBasicAuthConfigReq) ToEntity() *entity.HTTPBasicAuthConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPBasicAuthConfig{
		Enabled: r.Enabled,
		ID:      r.ID,
	}
}

//nolint:unparam
func (r *HTTPBasicAuthConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (r *HTTPBasicAuthConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateID(&r.ID, false, field+"id")...)
	return res
}

type HTTPClientConfigReq struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

func (r *HTTPClientConfigReq) ToEntity() *entity.HTTPClientConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPClientConfig{
		Enabled:        r.Enabled,
		MaxRequestBody: r.MaxRequestBody,
		MemRequestBody: r.MemRequestBody,
		AllowedIPs:     r.AllowedIPs,
	}
}

//nolint:unparam
func (r *HTTPClientConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	r.AllowedIPs = strings.FieldsFunc(strings.Join(r.AllowedIPs, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return nil
}

//nolint:unparam
func (r *HTTPClientConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	return res
}

type HTTPHeaderConfigReq struct {
	Enabled               bool              `json:"enabled"`
	AutoContentType       bool              `json:"autoContentType"`
	ToAddToRequests       map[string]string `json:"toAddToRequests"`
	ToRemoveFromRequests  []string          `json:"toRemoveFromRequests"`
	ToAddToResponses      map[string]string `json:"toAddToResponses"`
	ToRemoveFromResponses []string          `json:"toRemoveFromResponses"`
}

func (r *HTTPHeaderConfigReq) ToEntity() *entity.HTTPHeaderConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPHeaderConfig{
		Enabled:               r.Enabled,
		AutoContentType:       r.AutoContentType,
		ToAddToRequests:       r.ToAddToRequests,
		ToRemoveFromRequests:  r.ToRemoveFromRequests,
		ToAddToResponses:      r.ToAddToResponses,
		ToRemoveFromResponses: r.ToRemoveFromResponses,
	}
}

//nolint:unparam
func (r *HTTPHeaderConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (r *HTTPHeaderConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	return res
}

type HTTPCompressionConfigReq struct {
	Enabled              bool          `json:"enabled"`
	IncludedContentTypes []string      `json:"includedContentTypes"`
	ExcludedContentTypes []string      `json:"excludedContentTypes"`
	MinResponseBody      unit.DataSize `json:"minResponseBody"`
	DefaultEncoding      string        `json:"defaultEncoding"`
}

func (r *HTTPCompressionConfigReq) ToEntity() *entity.HTTPCompressionConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPCompressionConfig{
		Enabled:              r.Enabled,
		IncludedContentTypes: r.IncludedContentTypes,
		ExcludedContentTypes: r.ExcludedContentTypes,
		MinResponseBody:      r.MinResponseBody,
		DefaultEncoding:      r.DefaultEncoding,
	}
}

//nolint:unparam
func (r *HTTPCompressionConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	r.IncludedContentTypes = strings.FieldsFunc(strings.Join(r.IncludedContentTypes, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	r.ExcludedContentTypes = strings.FieldsFunc(strings.Join(r.ExcludedContentTypes, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return nil
}

//nolint:unparam
func (r *HTTPCompressionConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	return res
}

type HTTPRateLimitConfigReq struct {
	Enabled        bool              `json:"enabled"`
	Average        int               `json:"average"`
	Period         timeutil.Duration `json:"period"`
	Burst          int               `json:"burst"`
	MaxInFlightReq int               `json:"maxInFlightReq"`
}

func (r *HTTPRateLimitConfigReq) ToEntity() *entity.HTTPRateLimitConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPRateLimitConfig{
		Enabled:        r.Enabled,
		Average:        r.Average,
		Period:         r.Period,
		Burst:          r.Burst,
		MaxInFlightReq: r.MaxInFlightReq,
	}
}

//nolint:unparam
func (r *HTTPRateLimitConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (r *HTTPRateLimitConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	return res
}

type HTTPPathRewriteConfigReq struct {
	Enabled            bool   `json:"enabled"`
	PrefixAdd          string `json:"prefixAdd"`
	PrefixStrip        string `json:"prefixStrip"`
	PrefixStripIsRegex bool   `json:"prefixStripIsRegex"`
	PathReplace        string `json:"pathReplace"`
	PathReplaceIsRegex bool   `json:"pathReplaceIsRegex"`
	PathReplaceWith    string `json:"pathReplaceWith"`
}

func (r *HTTPPathRewriteConfigReq) ToEntity() *entity.HTTPPathRewriteConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPPathRewriteConfig{
		Enabled:            r.Enabled,
		PrefixAdd:          r.PrefixAdd,
		PrefixStrip:        r.PrefixStrip,
		PrefixStripIsRegex: r.PrefixStripIsRegex,
		PathReplace:        r.PathReplace,
		PathReplaceIsRegex: r.PathReplaceIsRegex,
		PathReplaceWith:    r.PathReplaceWith,
	}
}

//nolint:unparam
func (r *HTTPPathRewriteConfigReq) modifyRequest() error {
	if r == nil {
		return nil
	}
	if !r.PrefixStripIsRegex {
		prefixesStrip := strings.FieldsFunc(r.PrefixStrip, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		r.PrefixStrip = strings.Join(prefixesStrip, ",")
	}
	return nil
}

//nolint:unparam
func (r *HTTPPathRewriteConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	return res
}

type HTTPCircuitBreakerConfigReq struct {
	Enabled          bool              `json:"enabled"`
	Expression       string            `json:"expression"`
	CheckPeriod      timeutil.Duration `json:"checkPeriod"`
	FallbackDuration timeutil.Duration `json:"fallbackDuration"`
	RecoveryDuration timeutil.Duration `json:"recoveryDuration"`
	ResponseCode     int               `json:"responseCode"`
}

func (r *HTTPCircuitBreakerConfigReq) ToEntity() *entity.HTTPCircuitBreakerConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPCircuitBreakerConfig{
		Enabled:          r.Enabled,
		Expression:       r.Expression,
		CheckPeriod:      r.CheckPeriod,
		FallbackDuration: r.FallbackDuration,
		RecoveryDuration: r.RecoveryDuration,
		ResponseCode:     r.ResponseCode,
	}
}

//nolint:unparam
func (r *HTTPCircuitBreakerConfigReq) modifyRequest() error {
	return nil
}

//nolint:unparam
func (r *HTTPCircuitBreakerConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&r.Expression, true, 1, 1000, //nolint:mnd
		field+"expression")...)
	return res
}

type HTTPPathConfigReq struct {
	Enabled              bool                         `json:"enabled"`
	Path                 string                       `json:"path"`
	Mode                 base.HTTPPathMode            `json:"mode"`
	BasicAuth            *HTTPBasicAuthConfigReq      `json:"basicAuth"`
	ClientConfig         *HTTPClientConfigReq         `json:"clientConfig"`
	HeaderConfig         *HTTPHeaderConfigReq         `json:"headerConfig"`
	CompressionConfig    *HTTPCompressionConfigReq    `json:"compressionConfig"`
	RateLimitConfig      *HTTPRateLimitConfigReq      `json:"rateLimitConfig"`
	PathRewriteConfig    *HTTPPathRewriteConfigReq    `json:"pathRewriteConfig"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigReq `json:"circuitBreakerConfig"`
}

func (r *HTTPPathConfigReq) ToEntity() *entity.HTTPPathConfig {
	if r == nil {
		return nil
	}
	return &entity.HTTPPathConfig{
		Enabled:              r.Enabled,
		Path:                 r.Path,
		Mode:                 r.Mode,
		BasicAuth:            r.BasicAuth.ToEntity(),
		ClientConfig:         r.ClientConfig.ToEntity(),
		HeaderConfig:         r.HeaderConfig.ToEntity(),
		CompressionConfig:    r.CompressionConfig.ToEntity(),
		RateLimitConfig:      r.RateLimitConfig.ToEntity(),
		PathRewriteConfig:    r.PathRewriteConfig.ToEntity(),
		CircuitBreakerConfig: r.CircuitBreakerConfig.ToEntity(),
	}
}

//nolint:unparam
func (r *HTTPPathConfigReq) modifyRequest() error {
	if err := r.BasicAuth.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.ClientConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.HeaderConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.CompressionConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.RateLimitConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.PathRewriteConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	if err := r.CircuitBreakerConfig.modifyRequest(); err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

//nolint:unparam
func (r *HTTPPathConfigReq) validate(field string) (res []vld.Validator) {
	if r == nil || !r.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&r.Mode, true, base.AllHTTPPathModes, field+"mode")...)
	res = append(res, r.BasicAuth.validate(field+"basicAuth")...)
	res = append(res, r.ClientConfig.validate(field+"clientConfig")...)
	res = append(res, r.HeaderConfig.validate(field+"headerConfig")...)
	res = append(res, r.CompressionConfig.validate(field+"compressionConfig")...)
	res = append(res, r.RateLimitConfig.validate(field+"rateLimitConfig")...)
	res = append(res, r.PathRewriteConfig.validate(field+"pathRewriteConfig")...)
	res = append(res, r.CircuitBreakerConfig.validate(field+"circuitBreakerConfig")...)
	return res
}

func NewUpdateAppHttpSettingsReq() *UpdateAppHttpSettingsReq {
	return &UpdateAppHttpSettingsReq{}
}

func (req *UpdateAppHttpSettingsReq) ModifyRequest() error {
	for _, domainReq := range req.Domains {
		if err := domainReq.modifyRequest(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppHttpSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, vld.Slice(req.Domains).ForEach(
		func(r *DomainReq, index int, elemValidator vld.ItemValidator) {
			elemValidator.Validate(r.validate(fmt.Sprintf("domains[%d]", index))...)
		}))
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppHttpSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
