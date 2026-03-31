package traefikservice

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
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/htpasswd"
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

type TraefikConfig struct {
	TLS *TraefikTLS `yaml:"tls,omitempty"`
}

type TraefikTLS struct {
	Certificates []TraefikTLSCertificate `yaml:"certificates,omitempty"`
}

type TraefikTLSCertificate struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type AppConfigData struct {
	HttpSettings *entity.AppHttpSettings
	RefObjects   *entity.RefObjects

	confPath string
	confData *TraefikConfig
}

func (s *traefikService) ApplyAppConfig(
	ctx context.Context,
	app *entity.App,
	service *swarm.Service,
	data *AppConfigData,
) error {
	httpSettings := data.HttpSettings
	data.confPath = filepath.Join(config.Current.DataPathTraefikEtcDynamic(), app.Key+".yml")

	// 1. Calculate labels and TLS certs
	labels := make(map[string]string)
	traefikConfig := &TraefikConfig{}
	data.confData = traefikConfig

	hasCerts := false

	if httpSettings != nil && httpSettings.Enabled {
		labels["traefik.enable"] = labelValueTrue

		for _, domain := range httpSettings.Domains {
			s.collectDomainConfig(app, domain, labels, traefikConfig, data, &hasCerts)
		}
	}

	// 2. Apply Labels
	err := s.updateSwarmServiceLabels(service, labels)
	if err != nil {
		return err
	}

	// 3. Write or delete YAML file
	if hasCerts {
		err := s.writeConfigFile(data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		// Ensure file does not exist if no certs are needed
		_ = os.Remove(data.confPath)
	}

	return nil
}

func (s *traefikService) collectDomainConfig(
	app *entity.App,
	domain *entity.AppDomain,
	labels map[string]string,
	traefikConfig *TraefikConfig,
	data *AppConfigData,
	hasCerts *bool,
) {
	domainKey := sanitizeRouterNameReplacer.Replace(domain.Domain)
	if domainKey == "" {
		return
	}

	serviceName := fmt.Sprintf("%s-%s-svc", app.Key, domainKey)
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] =
		strconv.Itoa(domain.ContainerPort)
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.passhostheader", serviceName)] = labelValueTrue

	routerName := fmt.Sprintf("%s-%s-router", app.Key, domainKey)
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] =
		fmt.Sprintf("Host(`%s`)", domain.Domain)
	labels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = serviceName
	labels[fmt.Sprintf("traefik.http.routers.%s.tls", routerName)] = labelValueTrue
	labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "websecure"

	var middlewares []string

	// Force Https config
	s.createForceHttpsConfig(domain.ForceHttps, domainKey, labels, &middlewares)

	// Redirect config
	needRedirection := domain.DomainRedirect != "" && domain.Domain != domain.DomainRedirect
	s.createRedirectionConfig(needRedirection, domainKey, domain.Domain, domain.DomainRedirect, labels, &middlewares)

	// Basic auth config
	s.createBasicAuthConfig(domain.BasicAuth, data.RefObjects, domainKey, labels, &middlewares)

	// Client config
	s.createClientConfig(domain.ClientConfig, domainKey, labels, &middlewares)

	// Compression config
	s.createCompressionConfig(domain.CompressionConfig, domainKey, labels, &middlewares)

	// RateLimit config
	s.createRateLimitConfig(domain.RateLimitConfig, domainKey, labels, &middlewares)

	if len(middlewares) > 0 {
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] =
			strings.Join(middlewares, ",")
	}

	if !domain.ForceHttps {
		routerNameHttp := fmt.Sprintf("%s-%s-router-http", app.Key, domainKey)
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerNameHttp)] =
			fmt.Sprintf("Host(`%s`)", domain.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%s.service", routerNameHttp)] = serviceName
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerNameHttp)] = "web"
		if len(middlewares) > 0 {
			labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerNameHttp)] =
				strings.Join(middlewares, ",")
		}
	}

	// Paths config
	for pathIdx, pathCfg := range domain.Paths {
		s.collectPathConfig(domain, pathCfg, pathIdx, routerName, domainKey, serviceName,
			middlewares, labels, data)
	}

	if s.addTLSCertificate(traefikConfig, domain.SSLCert.ID) {
		*hasCerts = true
	}
}

func (s *traefikService) collectPathConfig(
	domain *entity.AppDomain,
	pathCfg *entity.HTTPPathConfig,
	pathIdx int,
	domainRouterName string,
	domainKey string,
	serviceName string,
	sharedMiddlewares []string,
	labels map[string]string,
	data *AppConfigData,
) {
	if pathCfg.Path == "" {
		return
	}

	pathMiddlewares := make([]string, len(sharedMiddlewares))
	copy(pathMiddlewares, sharedMiddlewares)
	pathRouterName := fmt.Sprintf("%s-path-%d", domainRouterName, pathIdx)
	pathMwPrefix := fmt.Sprintf("%s-path-%d", domainKey, pathIdx)

	// Basic auth config for path
	s.createBasicAuthConfig(pathCfg.BasicAuth, data.RefObjects, pathMwPrefix, labels, &pathMiddlewares)

	// Client config for path
	s.createClientConfig(pathCfg.ClientConfig, pathMwPrefix, labels, &pathMiddlewares)

	// RateLimit config for path
	s.createRateLimitConfig(pathCfg.RateLimitConfig, pathMwPrefix, labels, &pathMiddlewares)

	// Apply Path router labels
	var pathRule string
	if pathCfg.IsRegex {
		pathRule = fmt.Sprintf("Host(`%s`) && PathRegexp(`%s`)", domain.Domain, pathCfg.Path)
	} else {
		pathRule = fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", domain.Domain, pathCfg.Path)
	}

	labels[fmt.Sprintf("traefik.http.routers.%s.rule", pathRouterName)] = pathRule
	labels[fmt.Sprintf("traefik.http.routers.%s.service", pathRouterName)] = serviceName
	labels[fmt.Sprintf("traefik.http.routers.%s.tls", pathRouterName)] = labelValueTrue
	labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", pathRouterName)] = "websecure"
	if len(pathMiddlewares) > 0 {
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", pathRouterName)] =
			strings.Join(pathMiddlewares, ",")
	}

	if !domain.ForceHttps {
		pathRouterNameHttp := fmt.Sprintf("%s-http", pathRouterName)
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", pathRouterNameHttp)] = pathRule
		labels[fmt.Sprintf("traefik.http.routers.%s.service", pathRouterNameHttp)] = serviceName
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", pathRouterNameHttp)] = "web"
		if len(pathMiddlewares) > 0 {
			labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", pathRouterNameHttp)] =
				strings.Join(pathMiddlewares, ",")
		}
	}
}

func (s *traefikService) createForceHttpsConfig(
	forceHttps bool,
	domainKey string,
	labels map[string]string,
	middlewares *[]string,
) {
	if !forceHttps {
		return
	}
	mwName := fmt.Sprintf("%s-redirectscheme", domainKey)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.scheme", mwName)] = "https"
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectscheme.permanent", mwName)] = labelValueTrue
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *traefikService) createRedirectionConfig(
	redirect bool,
	domainKey string,
	domain, redirectTo string,
	labels map[string]string,
	middlewares *[]string,
) {
	if !redirect {
		return
	}
	mwName := fmt.Sprintf("%s-redirectregex", domainKey)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.regex", mwName)] =
		fmt.Sprintf("^https?://%s/(.*)", domain)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.replacement", mwName)] =
		fmt.Sprintf("https://%s/${1}", redirectTo)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.redirectregex.permanent", mwName)] = labelValueTrue
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *traefikService) createClientConfig(
	clientCfg *entity.HTTPClientConfig,
	domainKey string,
	labels map[string]string,
	middlewares *[]string,
) {
	if clientCfg == nil || !clientCfg.Enabled {
		return
	}
	mwName := fmt.Sprintf("%s-buffering", domainKey)
	if clientCfg.MaxRequestBodyBytes > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.buffering.maxrequestbodybytes", mwName)] =
			strconv.Itoa(clientCfg.MaxRequestBodyBytes)
	}
	if clientCfg.MemRequestBodyBytes > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.buffering.memrequestbodybytes", mwName)] =
			strconv.Itoa(clientCfg.MemRequestBodyBytes)
	}
	if clientCfg.MaxRequestBodyBytes == 0 && clientCfg.MemRequestBodyBytes == 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.buffering.maxrequestbodybytes", mwName)] = "0"
	}
	*middlewares = append(*middlewares, mwName+middlewareProvider)

	if len(clientCfg.AllowedIPs) > 0 {
		mwNameIp := fmt.Sprintf("%s-ipallowlist", domainKey)
		labels[fmt.Sprintf("traefik.http.middlewares.%s.ipallowlist.sourcerange", mwNameIp)] =
			strings.Join(clientCfg.AllowedIPs, ",")
		*middlewares = append(*middlewares, mwNameIp+middlewareProvider)
	}
}

func (s *traefikService) createBasicAuthConfig(
	basicAuth entity.ObjectID,
	refObjects *entity.RefObjects,
	domainKey string,
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
			mwName := fmt.Sprintf("%s-basicauth", domainKey)
			labels[fmt.Sprintf("traefik.http.middlewares.%s.basicauth.users", mwName)] =
				fmt.Sprintf("%s:%s", basicAuthConfig.Username, hashedPasswd)
			*middlewares = append(*middlewares, mwName+middlewareProvider)
		}
	}
}

func (s *traefikService) createCompressionConfig(
	compCfg *entity.HTTPCompressionConfig,
	domainKey string,
	labels map[string]string,
	middlewares *[]string,
) {
	if compCfg == nil || !compCfg.Enabled {
		return
	}
	mwName := fmt.Sprintf("%s-compress", domainKey)
	labels[fmt.Sprintf("traefik.http.middlewares.%s.compress", mwName)] = labelValueTrue
	if len(compCfg.ExcludedContentTypes) > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.excludedcontenttypes", mwName)] =
			strings.Join(compCfg.ExcludedContentTypes, ",")
	}
	if len(compCfg.IncludedContentTypes) > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.includedcontenttypes", mwName)] =
			strings.Join(compCfg.IncludedContentTypes, ",")
	}
	if compCfg.MinResponseBodyBytes > 0 {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.minresponsebodybytes", mwName)] =
			strconv.Itoa(compCfg.MinResponseBodyBytes)
	}
	if compCfg.DefaultEncoding != "" {
		labels[fmt.Sprintf("traefik.http.middlewares.%s.compress.defaultencoding", mwName)] =
			compCfg.DefaultEncoding
	}
	*middlewares = append(*middlewares, mwName+middlewareProvider)
}

func (s *traefikService) createRateLimitConfig(
	rlCfg *entity.HTTPRateLimitConfig,
	domainKey string,
	labels map[string]string,
	middlewares *[]string,
) {
	if rlCfg == nil || !rlCfg.Enabled {
		return
	}
	if rlCfg.Average > 0 || rlCfg.Burst > 0 || rlCfg.Period > 0 {
		mwName := fmt.Sprintf("%s-ratelimit", domainKey)
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

	if rlCfg.InFlightReqAmount > 0 {
		mwName := fmt.Sprintf("%s-inflightreq", domainKey)
		labels[fmt.Sprintf("traefik.http.middlewares.%s.inflightreq.amount", mwName)] =
			strconv.Itoa(rlCfg.InFlightReqAmount)
		*middlewares = append(*middlewares, mwName+middlewareProvider)
	}
}

func (s *traefikService) updateSwarmServiceLabels(
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
		if strings.HasPrefix(k, "traefik.") {
			delete(spec.Labels, k)
		}
	}
	// Apply new labels
	for k, v := range labels {
		spec.Labels[k] = v
	}

	return nil
}

func (s *traefikService) addTLSCertificate(
	traefikConfig *TraefikConfig,
	certID string,
) bool {
	if certID == "" {
		return false
	}

	certFile := filepath.Join(certsDir, certID+".crt")
	keyFile := filepath.Join(certsDir, certID+".key")

	if traefikConfig.TLS == nil {
		traefikConfig.TLS = &TraefikTLS{}
	}

	alreadyAdded := false
	for _, c := range traefikConfig.TLS.Certificates {
		if c.CertFile == certFile {
			alreadyAdded = true
			break
		}
	}
	if !alreadyAdded {
		traefikConfig.TLS.Certificates = append(traefikConfig.TLS.Certificates, TraefikTLSCertificate{
			CertFile: certFile,
			KeyFile:  keyFile,
		})
	}
	return true
}

func (s *traefikService) writeConfigFile(
	data *AppConfigData,
) error {
	yamlData, err := yaml.Marshal(data.confData)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = os.WriteFile(data.confPath, yamlData, defaultConfFileMode)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *traefikService) RemoveAppConfig(
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
	confPath := filepath.Join(config.Current.DataPathTraefikEtcDynamic(), app.Key+".yml")
	err := os.Remove(confPath)
	if err != nil && !os.IsNotExist(err) {
		return apperrors.Wrap(err)
	}

	return nil
}
