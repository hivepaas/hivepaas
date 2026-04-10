package traefikserviceimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/swarm"
	"gopkg.in/yaml.v3"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/htpasswd"
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
)

const (
	defaultConfFileMode = 0o755

	certsDir = "/etc/traefik/ssl/certs"

	labelValueTrue     = "true"
	middlewareProvider = "@swarm"
)

var (
	sanitizeRouterNameReplacer = strings.NewReplacer(".", "-", "_", "-")
)

type appConfigData struct {
	*traefikservice.AppConfigData
	app      *entity.App
	confData *AppTraefikConfig
	hasCerts bool
}

type AppTraefikConfig struct {
	TLS *AppTraefikTLS `yaml:"tls,omitempty"`
}

type AppTraefikTLS struct {
	Certificates []AppTraefikTLSCertificate `yaml:"certificates,omitempty"`
}

type AppTraefikTLSCertificate struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

func (s *service) ApplyAppConfig(
	ctx context.Context,
	app *entity.App,
	service *swarm.Service,
	cfgData *traefikservice.AppConfigData,
) error {
	data := &appConfigData{
		AppConfigData: cfgData,
		app:           app,
	}
	httpSettings := data.HttpSettings

	// 1. Calculate labels and TLS certs
	labels := make(map[string]string)
	traefikConfig := &AppTraefikConfig{}
	data.confData = traefikConfig

	if httpSettings != nil && httpSettings.ExposePublicly {
		labels["traefik.enable"] = labelValueTrue
		labels["traefik.swarm.network"] = base.NetworkGlobalRouting

		for i, domain := range httpSettings.Domains {
			s.collectDomainConfig(app, domain, i, labels, traefikConfig, data)
		}
	}

	// 2. Apply Labels
	err := s.updateSwarmServiceLabels(service, labels)
	if err != nil {
		return err
	}

	// 3. Write or delete YAML file
	if data.hasCerts {
		err := s.writeAppConfigFile(data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		// Ensure file does not exist if no certs are needed
		_ = os.Remove(app.TraefikConfigPath())
	}

	return nil
}

func (s *service) collectDomainConfig(
	app *entity.App,
	domain *entity.AppDomain,
	domainIndex int,
	labels map[string]string,
	traefikConfig *AppTraefikConfig,
	data *appConfigData,
) {
	appKey := sanitizeRouterNameReplacer.Replace(app.Key)
	domainKey := sanitizeRouterNameReplacer.Replace(domain.Domain)
	if domainKey == "" {
		return
	}

	// Service
	serviceName := fmt.Sprintf("svc-%s", appKey)
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] =
		strconv.Itoa(domain.ContainerPort)
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.passhostheader", serviceName)] = labelValueTrue

	// Main router
	routerName := fmt.Sprintf("router-%v", domainIndex)
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] =
		fmt.Sprintf("Host(`%s`)", domain.Domain)
	labels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = serviceName
	labels[fmt.Sprintf("traefik.http.routers.%s.tls", routerName)] = labelValueTrue
	labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "websecure"

	var middlewares []string

	// Force Https config
	s.createForceHttpsConfig(domain.ForceHttps, routerName, labels, &middlewares)

	// Redirect config
	needRedirection := domain.DomainRedirect != "" && domain.Domain != domain.DomainRedirect
	s.createRedirectionConfig(needRedirection, routerName, domain.Domain, domain.DomainRedirect, labels, &middlewares)

	// Basic auth config
	s.createBasicAuthConfig(domain.BasicAuth, data.RefObjects, routerName, labels, &middlewares)

	// Client config
	s.createClientConfig(domain.ClientConfig, routerName, labels, &middlewares)

	// Header config
	s.createHeaderConfig(domain.HeaderConfig, routerName, labels, &middlewares)

	// Compression config
	s.createCompressionConfig(domain.CompressionConfig, routerName, labels, &middlewares)

	// RateLimit config
	s.createRateLimitConfig(domain.RateLimitConfig, routerName, labels, &middlewares)

	if len(middlewares) > 0 {
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] =
			strings.Join(middlewares, ",")
	}

	// Paths config
	for pathIdx, pathCfg := range domain.Paths {
		s.collectPathConfig(domain, pathCfg, pathIdx, routerName, serviceName,
			middlewares, labels, data)
	}

	if s.addTLSCertificate(traefikConfig, domain.SSLCert.ID) {
		data.hasCerts = true
	}
}

func (s *service) collectPathConfig(
	domain *entity.AppDomain,
	pathCfg *entity.HTTPPathConfig,
	pathIdx int,
	domainRouterName string,
	serviceName string,
	sharedMiddlewares []string,
	labels map[string]string,
	data *appConfigData,
) {
	if pathCfg.Path == "" {
		return
	}

	// Apply Path router labels
	var pathRule string
	switch pathCfg.Mode { //nolint
	case base.HTTPPathModePrefix:
		pathRule = fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", domain.Domain, pathCfg.Path)
	case base.HTTPPathModeRegex:
		pathRule = fmt.Sprintf("Host(`%s`) && PathRegexp(`%s`)", domain.Domain, pathCfg.Path)
	default:
		pathRule = fmt.Sprintf("Host(`%s`) && Path(`%s`)", domain.Domain, pathCfg.Path)
	}

	pathRouterName := fmt.Sprintf("%s-path-%d", domainRouterName, pathIdx)
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", pathRouterName)] = pathRule
	labels[fmt.Sprintf("traefik.http.routers.%s.service", pathRouterName)] = serviceName
	labels[fmt.Sprintf("traefik.http.routers.%s.tls", pathRouterName)] = labelValueTrue
	labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", pathRouterName)] = "websecure"

	pathMiddlewares := make([]string, len(sharedMiddlewares))
	copy(pathMiddlewares, sharedMiddlewares)

	// Basic auth config for path
	s.createBasicAuthConfig(pathCfg.BasicAuth, data.RefObjects, pathRouterName, labels, &pathMiddlewares)

	// Client config for path
	s.createClientConfig(pathCfg.ClientConfig, pathRouterName, labels, &pathMiddlewares)

	// Header config for path
	s.createHeaderConfig(pathCfg.HeaderConfig, pathRouterName, labels, &pathMiddlewares)

	// RateLimit config for path
	s.createRateLimitConfig(pathCfg.RateLimitConfig, pathRouterName, labels, &pathMiddlewares)

	if len(pathMiddlewares) > 0 {
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", pathRouterName)] =
			strings.Join(pathMiddlewares, ",")
	}
}

func (s *service) createForceHttpsConfig(
	forceHttps bool,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if !forceHttps {
		return
	}

	// Listen to HTTP, then redirect to HTTPS
	labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "web,websecure"

	mwName := fmt.Sprintf("%s-forcehttps", routerName)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.scheme", mwName)] = "https"
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.permanent", mwName)] = labelValueTrue
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *service) createRedirectionConfig(
	redirect bool,
	routerName string,
	domain, redirectTo string,
	labels map[string]string,
	middlewares *[]string,
) {
	if !redirect {
		return
	}
	mwName := fmt.Sprintf("%s-redirect", routerName)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.regex", mwName)] =
		fmt.Sprintf("^https?://%s/(.*)", domain)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.replacement", mwName)] =
		fmt.Sprintf("https://%s/${1}", redirectTo)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.permanent", mwName)] = labelValueTrue
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *service) createClientConfig(
	clientCfg *entity.HTTPClientConfig,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if clientCfg == nil || !clientCfg.Enabled {
		return
	}
	mwName := fmt.Sprintf("%s-buffering", routerName)
	if clientCfg.MaxRequestBody > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.buffering.maxrequestbodybytes", mwName)] =
			strconv.FormatUint(uint64(clientCfg.MaxRequestBody), 10)
	}
	if clientCfg.MemRequestBody > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.buffering.memrequestbodybytes", mwName)] =
			strconv.FormatUint(uint64(clientCfg.MemRequestBody), 10)
	}
	*middlewares = append(*middlewares, mwName+middlewareProvider)

	if len(clientCfg.AllowedIPs) > 0 {
		mwNameIp := fmt.Sprintf("%s-allowed-ips", routerName)
		labels[fmt.Sprintf("traefik.http.middlewares.%s.ipallowlist.sourcerange", mwNameIp)] =
			strings.Join(clientCfg.AllowedIPs, ",")
		*middlewares = append(*middlewares, mwNameIp+middlewareProvider)
	}
}

func (s *service) createHeaderConfig(
	headerCfg *entity.HTTPHeaderConfig,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if headerCfg == nil || (len(headerCfg.ToAddToRequests) == 0 && len(headerCfg.ToRemoveFromRequests) == 0) {
		return
	}
	mwName := fmt.Sprintf("%s-headers", routerName)

	for k, v := range headerCfg.ToAddToRequests {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.headers.customrequestheaders.%s", mwName, k)] = v
	}
	for k, v := range headerCfg.ToAddToResponses {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.headers.customresponseheaders.%s", mwName, k)] = v
	}

	for _, k := range headerCfg.ToRemoveFromRequests {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.headers.customrequestheaders.%s", mwName, k)] = ""
	}
	for _, k := range headerCfg.ToRemoveFromResponses {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.headers.customresponseheaders.%s", mwName, k)] = ""
	}

	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *service) createBasicAuthConfig(
	basicAuth entity.ObjectID,
	refObjects *entity.RefObjects,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if basicAuth.ID == "" {
		return
	}
	if s := refObjects.RefSettings[basicAuth.ID]; s != nil {
		basicAuthConfig := s.MustAsBasicAuth()
		hashedPasswd, err := htpasswd.HashPassword(basicAuthConfig.Password.MustGetPlain())
		if err == nil {
			mwName := fmt.Sprintf("%s-basicauth", routerName)
			labels[fmt.Sprintf("traefik.http.middlewares.%s.basicauth.users", mwName)] =
				fmt.Sprintf("%s:%s", basicAuthConfig.Username, hashedPasswd)
			*middlewares = append(*middlewares, mwName+middlewareProvider)
		}
	}
}

func (s *service) createCompressionConfig(
	compCfg *entity.HTTPCompressionConfig,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if compCfg == nil || !compCfg.Enabled {
		return
	}
	mwName := fmt.Sprintf("%s-compress", routerName)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.compress", mwName)] = labelValueTrue
	if len(compCfg.ExcludedContentTypes) > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.excludedcontenttypes", mwName)] =
			strings.Join(compCfg.ExcludedContentTypes, ",")
	}
	if len(compCfg.IncludedContentTypes) > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.includedcontenttypes", mwName)] =
			strings.Join(compCfg.IncludedContentTypes, ",")
	}
	if compCfg.MinResponseBody > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.minresponsebodybytes", mwName)] =
			strconv.FormatUint(uint64(compCfg.MinResponseBody), 10)
	}
	if compCfg.DefaultEncoding != "" {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.defaultencoding", mwName)] =
			compCfg.DefaultEncoding
	}
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *service) createRateLimitConfig(
	rlCfg *entity.HTTPRateLimitConfig,
	routerName string,
	labels map[string]string,
	middlewares *[]string,
) {
	if rlCfg == nil || !rlCfg.Enabled {
		return
	}
	if rlCfg.Average > 0 || rlCfg.Burst > 0 || rlCfg.Period > 0 {
		mwName := fmt.Sprintf("%s-ratelimit", routerName)
		if rlCfg.Average > 0 {
			labels[fmt.Sprintf("traefik.http.middlewares.%s.ratelimit.average", mwName)] =
				strconv.Itoa(rlCfg.Average)
		}
		if rlCfg.Burst > 0 {
			labels[fmt.Sprintf("traefik.http.middlewares.%s.ratelimit.burst", mwName)] =
				strconv.Itoa(rlCfg.Burst)
		}
		if rlCfg.Period > 0 {
			labels[fmt.Sprintf("traefik.http.middlewares.%s.ratelimit.period", mwName)] =
				rlCfg.Period.ToDuration().String()
		}
		*middlewares = append(*middlewares, mwName+middlewareProvider)
	}

	if rlCfg.MaxInFlightReq > 0 {
		mwName := fmt.Sprintf("%s-inflightreq", routerName)
		labels[fmt.Sprintf("traefik.http.middlewares.%s.inflightreq.amount", mwName)] =
			strconv.Itoa(rlCfg.MaxInFlightReq)
		*middlewares = append(*middlewares, mwName+middlewareProvider)
	}
}

func (s *service) updateSwarmServiceLabels(
	service *swarm.Service,
	labels map[string]string,
) error {
	if service == nil {
		return nil
	}
	spec := &service.Spec
	if spec.Labels == nil {
		spec.Labels = make(map[string]string)
	}
	// Clean old traefik labels
	for k := range spec.Labels {
		if strings.HasPrefix(k, "traefik.") && !strings.Contains(k, ".x-custom-") {
			delete(spec.Labels, k)
		}
	}
	// Apply new labels
	for k, v := range labels {
		spec.Labels[k] = v
	}

	return nil
}

func (s *service) addTLSCertificate(
	traefikConfig *AppTraefikConfig,
	certID string,
) bool {
	if certID == "" {
		return false
	}

	certFile := filepath.Join(certsDir, certID+".crt")
	keyFile := filepath.Join(certsDir, certID+".key")

	if traefikConfig.TLS == nil {
		traefikConfig.TLS = &AppTraefikTLS{}
	}

	alreadyAdded := false
	for _, c := range traefikConfig.TLS.Certificates {
		if c.CertFile == certFile {
			alreadyAdded = true
			break
		}
	}
	if !alreadyAdded {
		traefikConfig.TLS.Certificates = append(traefikConfig.TLS.Certificates, AppTraefikTLSCertificate{
			CertFile: certFile,
			KeyFile:  keyFile,
		})
	}
	return true
}

func (s *service) writeAppConfigFile(
	data *appConfigData,
) error {
	yamlData, err := yaml.Marshal(data.confData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = os.WriteFile(data.app.TraefikConfigPath(), yamlData, defaultConfFileMode)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) RemoveAppConfig(
	_ context.Context,
	app *entity.App,
	service *swarm.Service,
) error {
	// Clean from Swarm Service
	if service != nil && service.Spec.Labels != nil {
		for k := range service.Spec.Labels {
			if strings.HasPrefix(k, "traefik.") {
				delete(service.Spec.Labels, k)
			}
		}
	}

	// Clean file
	err := os.Remove(app.TraefikConfigPath())
	if err != nil && !os.IsNotExist(err) {
		return apperrors.Wrap(err)
	}

	return nil
}
