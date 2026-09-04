package appsettingsdto

import (
	"fmt"
	"strings"
	"unicode"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/traefik"
)

type UpdateAppRoutingSettingsReq struct {
	ProjectID      string       `json:"-"`
	ProjectEnvID   string       `json:"-"`
	AppID          string       `json:"-"`
	Port           int          `json:"port"`
	ExposePublicly bool         `json:"exposePublicly"`
	Domains        []*DomainReq `json:"domains"`
	UpdateVer      int          `json:"updateVer"`
}

func (req *UpdateAppRoutingSettingsReq) ToEntity() *entity.AppRoutingSettings {
	return &entity.AppRoutingSettings{
		Port:           req.Port,
		ExposePublicly: req.ExposePublicly,
		Domains: gofn.MapSlice(req.Domains, func(r *DomainReq) *entity.AppDomain {
			return r.ToEntity()
		}),
	}
}

type DomainReq struct {
	Enabled        bool                 `json:"enabled"`
	Domain         string               `json:"domain"`
	Protocol       base.NetworkProtocol `json:"protocol"`
	ContainerPort  int                  `json:"containerPort"`
	SSLCert        basedto.ObjectIDReq  `json:"sslCert"`
	TLSPassthrough bool                 `json:"tlsPassthrough"`

	// HTTP (layer 7) configuration
	DomainRedirect       string                       `json:"domainRedirect"`
	ForceHttps           bool                         `json:"forceHttps"`
	LBConfig             *HTTPLBConfigReq             `json:"lbConfig"`
	BasicAuth            *HTTPBasicAuthConfigReq      `json:"basicAuth"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigReq `json:"circuitBreakerConfig"`
	ClientConfig         *HTTPClientConfigReq         `json:"clientConfig"`
	CompressionConfig    *HTTPCompressionConfigReq    `json:"compressionConfig"`
	HeaderConfig         *HTTPHeaderConfigReq         `json:"headerConfig"`
	PathRewriteConfig    *HTTPPathRewriteConfigReq    `json:"pathRewriteConfig"`
	RateLimitConfig      *HTTPRateLimitConfigReq      `json:"rateLimitConfig"`
	WebsocketConfig      *HTTPWebsocketConfigReq      `json:"websocketConfig"`
	Paths                []*HTTPPathConfigReq         `json:"paths"`
}

func (req *DomainReq) ToEntity() *entity.AppDomain {
	proto := req.Protocol
	if proto == "" {
		proto = base.NetworkProtocolHTTP
	}
	return &entity.AppDomain{
		Enabled:              req.Enabled,
		Domain:               req.Domain,
		Protocol:             proto,
		ContainerPort:        req.ContainerPort,
		SSLCert:              entity.ObjectID{ID: req.SSLCert.ID},
		TLSPassthrough:       req.TLSPassthrough,
		DomainRedirect:       req.DomainRedirect,
		ForceHttps:           req.ForceHttps,
		LBConfig:             req.LBConfig.ToEntity(),
		BasicAuth:            req.BasicAuth.ToEntity(),
		CircuitBreakerConfig: req.CircuitBreakerConfig.ToEntity(),
		ClientConfig:         req.ClientConfig.ToEntity(),
		CompressionConfig:    req.CompressionConfig.ToEntity(),
		HeaderConfig:         req.HeaderConfig.ToEntity(),
		PathRewriteConfig:    req.PathRewriteConfig.ToEntity(),
		RateLimitConfig:      req.RateLimitConfig.ToEntity(),
		WebsocketConfig:      req.WebsocketConfig.ToEntity(),
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
	if req.Protocol == "" {
		req.Protocol = base.NetworkProtocolHTTP
	}
	req.Domain = strings.ToLower(strings.TrimSpace(req.Domain))
	if req.Protocol == base.NetworkProtocolHTTP {
		if err := req.LBConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.BasicAuth.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.CircuitBreakerConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.ClientConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.CompressionConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.HeaderConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.PathRewriteConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.RateLimitConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		if err := req.WebsocketConfig.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
		for _, pathReq := range req.Paths {
			if err := pathReq.modifyRequest(); err != nil {
				return hperrors.Wrap(err)
			}
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
	res = append(res, basedto.ValidateStrIn(&req.Protocol, false, base.AllRoutingProtocols, field+"protocol")...)
	res = append(res, basedto.ValidateDomain(&req.Domain, true, base.DomainNameMaxLen,
		false, field+"domain")...)
	res = append(res, basedto.ValidatePort(&req.ContainerPort, false, 1, field+"containerPort")...)

	if req.Protocol == base.NetworkProtocolHTTP {
		res = append(res, basedto.ValidateDomain(&req.DomainRedirect, false, base.DomainNameMaxLen,
			false, field+"domainRedirect")...)
		res = append(res, req.LBConfig.validate(field+"lbConfig")...)
		res = append(res, req.BasicAuth.validate(field+"basicAuth")...)
		res = append(res, req.CircuitBreakerConfig.validate(field+"circuitBreakerConfig")...)
		res = append(res, req.ClientConfig.validate(field+"clientConfig")...)
		res = append(res, req.CompressionConfig.validate(field+"compressionConfig")...)
		res = append(res, req.HeaderConfig.validate(field+"headerConfig")...)
		res = append(res, req.PathRewriteConfig.validate(field+"pathRewriteConfig")...)
		res = append(res, req.RateLimitConfig.validate(field+"rateLimitConfig")...)
		res = append(res, req.WebsocketConfig.validate(field+"websocketConfig")...)
		for i, pathReq := range req.Paths {
			res = append(res, pathReq.validate(field+fmt.Sprintf("paths[%v]", i))...)
		}
	}
	return res
}

type HTTPLBConfigReq struct {
	Strategy traefik.LBStrategy `json:"strategy"`
}

func (req *HTTPLBConfigReq) ToEntity() *entity.HTTPLBConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPLBConfig{
		Strategy: req.Strategy,
	}
}

//nolint:unparam
func (req *HTTPLBConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (req *HTTPLBConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Strategy, false, traefik.AllLBStrategies, field+"strategy")...)
	return res
}

type HTTPBasicAuthConfigReq struct {
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

func (req *HTTPBasicAuthConfigReq) ToEntity() *entity.HTTPBasicAuthConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPBasicAuthConfig{
		Enabled: req.Enabled,
		ID:      req.ID,
	}
}

//nolint:unparam
func (req *HTTPBasicAuthConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (req *HTTPBasicAuthConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateID(&req.ID, false, field+"id")...)
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

func (req *HTTPCircuitBreakerConfigReq) ToEntity() *entity.HTTPCircuitBreakerConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPCircuitBreakerConfig{
		Enabled:          req.Enabled,
		Expression:       req.Expression,
		CheckPeriod:      req.CheckPeriod,
		FallbackDuration: req.FallbackDuration,
		RecoveryDuration: req.RecoveryDuration,
		ResponseCode:     req.ResponseCode,
	}
}

//nolint:unparam
func (req *HTTPCircuitBreakerConfigReq) modifyRequest() error {
	return nil
}

//nolint:unparam
func (req *HTTPCircuitBreakerConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Expression, true, 1, 1000, //nolint:mnd
		field+"expression")...)
	return res
}

type HTTPClientConfigReq struct {
	Enabled        bool          `json:"enabled"`
	MaxRequestBody unit.DataSize `json:"maxRequestBody"`
	MemRequestBody unit.DataSize `json:"memRequestBody"`
	AllowedIPs     []string      `json:"allowedIPs"`
}

func (req *HTTPClientConfigReq) ToEntity() *entity.HTTPClientConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPClientConfig{
		Enabled:        req.Enabled,
		MaxRequestBody: req.MaxRequestBody,
		MemRequestBody: req.MemRequestBody,
		AllowedIPs:     req.AllowedIPs,
	}
}

//nolint:unparam
func (req *HTTPClientConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.AllowedIPs = strings.FieldsFunc(strings.Join(req.AllowedIPs, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return nil
}

//nolint:unparam
func (req *HTTPClientConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
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

func (req *HTTPCompressionConfigReq) ToEntity() *entity.HTTPCompressionConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPCompressionConfig{
		Enabled:              req.Enabled,
		IncludedContentTypes: req.IncludedContentTypes,
		ExcludedContentTypes: req.ExcludedContentTypes,
		MinResponseBody:      req.MinResponseBody,
		DefaultEncoding:      req.DefaultEncoding,
	}
}

//nolint:unparam
func (req *HTTPCompressionConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.IncludedContentTypes = strings.FieldsFunc(strings.Join(req.IncludedContentTypes, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	req.ExcludedContentTypes = strings.FieldsFunc(strings.Join(req.ExcludedContentTypes, ","), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	return nil
}

//nolint:unparam
func (req *HTTPCompressionConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
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

func (req *HTTPHeaderConfigReq) ToEntity() *entity.HTTPHeaderConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPHeaderConfig{
		Enabled:               req.Enabled,
		AutoContentType:       req.AutoContentType,
		ToAddToRequests:       req.ToAddToRequests,
		ToRemoveFromRequests:  req.ToRemoveFromRequests,
		ToAddToResponses:      req.ToAddToResponses,
		ToRemoveFromResponses: req.ToRemoveFromResponses,
	}
}

//nolint:unparam
func (req *HTTPHeaderConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (req *HTTPHeaderConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
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

func (req *HTTPPathRewriteConfigReq) ToEntity() *entity.HTTPPathRewriteConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPPathRewriteConfig{
		Enabled:            req.Enabled,
		PrefixAdd:          req.PrefixAdd,
		PrefixStrip:        req.PrefixStrip,
		PrefixStripIsRegex: req.PrefixStripIsRegex,
		PathReplace:        req.PathReplace,
		PathReplaceIsRegex: req.PathReplaceIsRegex,
		PathReplaceWith:    req.PathReplaceWith,
	}
}

//nolint:unparam
func (req *HTTPPathRewriteConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	if !req.PrefixStripIsRegex {
		prefixesStrip := strings.FieldsFunc(req.PrefixStrip, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		req.PrefixStrip = strings.Join(prefixesStrip, ",")
	}
	return nil
}

//nolint:unparam
func (req *HTTPPathRewriteConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
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

func (req *HTTPRateLimitConfigReq) ToEntity() *entity.HTTPRateLimitConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPRateLimitConfig{
		Enabled:        req.Enabled,
		Average:        req.Average,
		Period:         req.Period,
		Burst:          req.Burst,
		MaxInFlightReq: req.MaxInFlightReq,
	}
}

//nolint:unparam
func (req *HTTPRateLimitConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	return nil
}

//nolint:unparam
func (req *HTTPRateLimitConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	return res
}

type HTTPWebsocketConfigReq struct {
	Enabled bool `json:"enabled"`
}

func (req *HTTPWebsocketConfigReq) ToEntity() *entity.HTTPWebsocketConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPWebsocketConfig{
		Enabled: req.Enabled,
	}
}

//nolint:unparam
func (req *HTTPWebsocketConfigReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.Enabled = true // NOTE: in BE side, websocket is always enabled
	return nil
}

//nolint:unparam
func (req *HTTPWebsocketConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	return res
}

type HTTPPathConfigReq struct {
	Enabled              bool                         `json:"enabled"`
	Path                 string                       `json:"path"`
	Mode                 base.HTTPPathMode            `json:"mode"`
	BasicAuth            *HTTPBasicAuthConfigReq      `json:"basicAuth"`
	CircuitBreakerConfig *HTTPCircuitBreakerConfigReq `json:"circuitBreakerConfig"`
	ClientConfig         *HTTPClientConfigReq         `json:"clientConfig"`
	CompressionConfig    *HTTPCompressionConfigReq    `json:"compressionConfig"`
	HeaderConfig         *HTTPHeaderConfigReq         `json:"headerConfig"`
	PathRewriteConfig    *HTTPPathRewriteConfigReq    `json:"pathRewriteConfig"`
	RateLimitConfig      *HTTPRateLimitConfigReq      `json:"rateLimitConfig"`
	WebsocketConfig      *HTTPWebsocketConfigReq      `json:"websocketConfig"`
}

func (req *HTTPPathConfigReq) ToEntity() *entity.HTTPPathConfig {
	if req == nil {
		return nil
	}
	return &entity.HTTPPathConfig{
		Enabled:              req.Enabled,
		Path:                 req.Path,
		Mode:                 req.Mode,
		BasicAuth:            req.BasicAuth.ToEntity(),
		CircuitBreakerConfig: req.CircuitBreakerConfig.ToEntity(),
		ClientConfig:         req.ClientConfig.ToEntity(),
		CompressionConfig:    req.CompressionConfig.ToEntity(),
		HeaderConfig:         req.HeaderConfig.ToEntity(),
		PathRewriteConfig:    req.PathRewriteConfig.ToEntity(),
		RateLimitConfig:      req.RateLimitConfig.ToEntity(),
		WebsocketConfig:      req.WebsocketConfig.ToEntity(),
	}
}

//nolint:unparam
func (req *HTTPPathConfigReq) modifyRequest() error {
	if err := req.BasicAuth.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.CircuitBreakerConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.ClientConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.CompressionConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.HeaderConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.PathRewriteConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.RateLimitConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	if err := req.WebsocketConfig.modifyRequest(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

//nolint:unparam
func (req *HTTPPathConfigReq) validate(field string) (res []vld.Validator) {
	if req == nil || !req.Enabled {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Mode, true, base.AllHTTPPathModes, field+"mode")...)
	res = append(res, req.BasicAuth.validate(field+"basicAuth")...)
	res = append(res, req.CircuitBreakerConfig.validate(field+"circuitBreakerConfig")...)
	res = append(res, req.ClientConfig.validate(field+"clientConfig")...)
	res = append(res, req.CompressionConfig.validate(field+"compressionConfig")...)
	res = append(res, req.HeaderConfig.validate(field+"headerConfig")...)
	res = append(res, req.PathRewriteConfig.validate(field+"pathRewriteConfig")...)
	res = append(res, req.RateLimitConfig.validate(field+"rateLimitConfig")...)
	res = append(res, req.WebsocketConfig.validate(field+"websocketConfig")...)
	return res
}

func NewUpdateAppRoutingSettingsReq() *UpdateAppRoutingSettingsReq {
	return &UpdateAppRoutingSettingsReq{}
}

func (req *UpdateAppRoutingSettingsReq) ModifyRequest() error {
	activePort := 0
	for _, domainReq := range req.Domains {
		if activePort == 0 && domainReq.Enabled {
			activePort = domainReq.ContainerPort
		}
		if err := domainReq.modifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if req.Port == 0 {
		req.Port = activePort
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppRoutingSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidatePort(&req.Port, false, 1, "port")...)
	validators = append(validators, vld.Slice(req.Domains).ForEach(
		func(r *DomainReq, index int, elemValidator vld.ItemValidator) {
			elemValidator.Validate(r.validate(fmt.Sprintf("domains[%d]", index))...)
		}))
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppRoutingSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
