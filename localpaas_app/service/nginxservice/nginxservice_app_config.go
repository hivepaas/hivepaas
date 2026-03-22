package nginxservice

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/services/nginx"
)

const (
	defaultConfBuf      = 10 * 1024 // 10KB
	defaultConfFileMode = 0o755
)

const (
	nginxShareDir = "/usr/share/nginx" // dir path within container of nginx

	certsBaseDir = nginxShareDir + "/certs"
	certsMainDir = certsBaseDir + "/main"
	certsFakeDir = certsBaseDir + "/fake" // self-signed

	basicAuthDir = nginxShareDir + "/basic-auth"
)

var (
	appConfTemplate = func() []byte {
		data, err := os.ReadFile("config/nginx/app.conf.template")
		if err != nil {
			panic(err)
		}
		return data
	}()

	httpsRedirectConfTemplate = func() []byte {
		data, err := os.ReadFile("config/nginx/https_redirect.conf.template")
		if err != nil {
			panic(err)
		}
		return data
	}()

	domainRedirectConfTemplate = func() []byte {
		data, err := os.ReadFile("config/nginx/domain_redirect.conf.template")
		if err != nil {
			panic(err)
		}
		return data
	}()
)

const (
	clientConfDefault = `
		client_max_body_size 500m;
	`

	gzipConfDefault = `
		gzip on;
		gzip_disable "MSIE [1-6]\.";
		gzip_vary on;
		gzip_proxied any;
		gzip_comp_level 6;
		gzip_min_length 10240;
		gzip_buffers 16 8k;
		gzip_types
			text/css
			text/plain
			text/javascript
			application/javascript
			application/json
			application/x-javascript
			application/xml
			application/xml+rss
			application/xhtml+xml
			application/x-font-ttf
			application/x-font-opentype
			application/vnd.ms-fontobject
			image/svg+xml
			image/x-icon
			application/rss+xml
			application/atom_xml;
	`

	proxyHeaderConfDefault = `
		proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
	`

	websocketConfDefault = `
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_http_version 1.1;
	`

	locationConfDefault = `
        # @lp_section proxy_header
        # @lp_section_end proxy_header

        # @lp_section websocket
        # @lp_section_end websocket

        # @lp_section limit_req
        # @lp_section_end limit_req

        # @lp_section auth_basic
        # @lp_section_end auth_basic

        # @lp_section custom
        # @lp_section_end custom
	`
)

type AppConfigData struct {
	HttpSettings *entity.AppHttpSettings
	RefObjects   *entity.RefObjects

	confPath string
	confData map[string]*domainConfig
}

type domainConfig struct {
	App    *entity.App
	Domain *entity.AppDomain
	Reset  bool

	HttpsRedirectConf  *nginx.Block
	DomainRedirectConf *nginx.Block
	MainConf           *nginx.Block
}

func (s *nginxService) ApplyAppConfig(
	ctx context.Context,
	app *entity.App,
	data *AppConfigData,
) (err error) {
	httpSettings := data.HttpSettings
	data.confPath = filepath.Join(config.Current.DataPathNginxEtcConf(), app.Key+".conf")

	// Not enabled, delete the config file, then return
	if httpSettings == nil || !httpSettings.Enabled {
		return s.RemoveAppConfig(ctx, app)
	}

	data.confData = make(map[string]*domainConfig, len(data.HttpSettings.Domains))
	for _, domain := range data.HttpSettings.Domains {
		data.confData[domain.Domain] = &domainConfig{
			App:    app,
			Domain: domain,
			Reset:  data.HttpSettings.Reset,
		}
	}

	for _, domainConf := range data.confData {
		err = s.buildHttpsRedirectionConfig(domainConf)
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = s.buildDomainRedirectionConfig(domainConf)
		if err != nil {
			return apperrors.Wrap(err)
		}
		err = s.buildMainConfig(domainConf)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	err = s.writeConfigFile(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *nginxService) writeConfigFile(
	ctx context.Context,
	data *AppConfigData,
) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0, defaultConfBuf))
	for _, domainConf := range data.confData {
		if domainConf.HttpsRedirectConf != nil {
			err = nginxBuildConfDefault(domainConf.HttpsRedirectConf.AsConfig(), buf)
			if err != nil {
				return apperrors.New(err).WithMsgLog("failed to build https_redirect conf")
			}
		}
		if domainConf.DomainRedirectConf != nil {
			err = nginxBuildConfDefault(domainConf.DomainRedirectConf.AsConfig(), buf)
			if err != nil {
				return apperrors.New(err).WithMsgLog("failed to build domain redirect conf")
			}
		}
		if domainConf.MainConf != nil {
			err = nginxBuildConfDefault(domainConf.MainConf.AsConfig(), buf)
			if err != nil {
				return apperrors.New(err).WithMsgLog("failed to build main server conf")
			}
		}
	}

	err = os.WriteFile(data.confPath, buf.Bytes(), defaultConfFileMode)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// TODO: check config correctness with nginx -t

	// Requests nginx to reload the config files
	err = s.ReloadNginxConfig(ctx, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *nginxService) buildHttpsRedirectionConfig(
	domainConf *domainConfig,
) error {
	domain := domainConf.Domain
	if !domain.ForceHttps {
		return nil
	}

	conf, err := nginxParseConfDefault(string(httpsRedirectConfTemplate))
	if err != nil {
		return apperrors.Wrap(err)
	}
	serverBlock := gofn.FirstOr(conf.BlocksByName("server", 1), nil)
	domainConf.HttpsRedirectConf = serverBlock

	confSetServerName(serverBlock, domain.Domain)

	return nil
}

func (s *nginxService) buildDomainRedirectionConfig(
	domainConf *domainConfig,
) error {
	domain := domainConf.Domain
	needRedirection := domain.DomainRedirect != "" && domain.Domain != domain.DomainRedirect

	if !needRedirection {
		domainConf.DomainRedirectConf = nil
		return nil
	}

	conf, err := nginxParseConfDefault(string(domainRedirectConfTemplate))
	if err != nil {
		return apperrors.Wrap(err)
	}
	serverBlock := gofn.FirstOr(conf.BlocksByName("server", 1), nil)
	domainConf.DomainRedirectConf = serverBlock

	confSetServerName(serverBlock, domain.Domain)
	confSetSSLCert(serverBlock, domain.SSLCert.ID)

	serverBlock.SetVariable("$lp_redirect_to", domain.DomainRedirect)

	return nil
}

func (s *nginxService) buildMainConfig(
	domainConf *domainConfig,
) error {
	domain := domainConf.Domain
	if domainConf.DomainRedirectConf != nil {
		return nil
	}

	conf, err := nginxParseConfDefault(string(appConfTemplate))
	if err != nil {
		return apperrors.Wrap(err)
	}
	serverBlock := gofn.FirstOr(conf.BlocksByName("server", 1), nil)
	domainConf.MainConf = serverBlock

	confSetServerName(serverBlock, domain.Domain)
	confSetSSLCert(serverBlock, domain.SSLCert.ID)
	confSetUpstream(serverBlock, domainConf.App.Key, domain.ContainerPort)

	serverConfig := gofn.Coalesce(domain.NginxSettings, &entity.NginxSettings{})

	err = confSetClient(serverBlock, serverConfig.ClientConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetGzip(serverBlock, serverConfig.GzipConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetBasicAuth(serverBlock, domain.BasicAuth.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetCustom(serverBlock, serverConfig.CustomConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	var allLocations []*entity.NginxLocationBlock
	for _, location := range serverConfig.Locations {
		if location.Location == "/" {
			allLocations = append([]*entity.NginxLocationBlock{location}, allLocations...)
		} else {
			allLocations = append(allLocations, location)
		}
	}
	if len(allLocations) == 0 || allLocations[0].Location != "/" {
		allLocations = append([]*entity.NginxLocationBlock{
			{
				Location:          "/",
				ProxyHeaderConfig: "default",
			},
		}, allLocations...)
	}
	for _, location := range allLocations {
		err = s.buildMainLocationConfig(domainConf, serverBlock, location)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *nginxService) buildMainLocationConfig(
	_ *domainConfig,
	serverBlock *nginx.Block,
	locationConf *entity.NginxLocationBlock,
) error {
	locationBlock := serverBlock.GetBlock("location", []string{locationConf.Location}, true)
	if locationBlock == nil {
		locationBlock = nginx.NewLocationBlock([]string{locationConf.Location})
		defaultConf, err := nginxParseConfDefault(locationConfDefault)
		if err != nil {
			return apperrors.Wrap(err)
		}
		locationBlock.AddDirectives(defaultConf.AllDirectives()...)
		serverBlock.AddBlock(locationBlock)
	}

	err := confSetProxyHeader(locationBlock, locationConf.ProxyHeaderConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetWebsocket(locationBlock, locationConf.WebsocketConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetBasicAuth(locationBlock, locationConf.BasicAuth.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetLimitReq(locationBlock, locationConf.LimitReqConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}
	err = confSetCustom(locationBlock, locationConf.CustomConfig)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func confSetServerName(serverBlock *nginx.Block, serverName string) {
	lpServer := serverBlock.GetComment("@lp_server ", true)
	if lpServer != nil {
		lpServer.Comment = new(strings.Replace(*lpServer.Comment, "server_name", serverName, 1))
	}
	serverBlock.SetDirectiveArgs("server_name", []string{serverName}, 1)
}

func confSetSSLCert(serverBlock *nginx.Block, sslID string) {
	section := "@lp_section ssl"
	certFile, keyFile := getSSLFilePaths(sslID)
	serverBlock.AddDirectivesInSection(section,
		nginx.NewDirective("ssl_certificate", []string{certFile}),
		nginx.NewDirective("ssl_certificate_key", []string{keyFile}),
	)
}

func confSetUpstream(serverBlock *nginx.Block, key string, containerPort int) {
	upstream := fmt.Sprintf("http://%s:%d", key, containerPort)
	_ = serverBlock.SetVariable("$upstream", upstream)
}

func confSetClient(serverBlock *nginx.Block, config string) error {
	section := "@lp_section client"
	if config == "" {
		serverBlock.RemoveSectionComments(section)
		return nil
	}
	if config == "default" { //nolint:gocritic,goconst
		config = clientConfDefault
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		serverBlock.AddDirectivesInSection(section, directives...)
	} else {
		serverBlock.RemoveSectionComments(section)
	}
	return nil
}

func confSetGzip(serverBlock *nginx.Block, config string) error {
	section := "@lp_section gzip"
	if config == "" {
		serverBlock.RemoveSectionComments(section)
		return nil
	}
	switch config { //nolint:gocritic
	case "default":
		config = gzipConfDefault
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		serverBlock.AddDirectivesInSection(section, directives...)
	} else {
		serverBlock.RemoveSectionComments(section)
	}
	return nil
}

func confSetWebsocket(block *nginx.Block, config string) error {
	section := "@lp_section websocket"
	if config == "" {
		block.RemoveSectionComments(section)
		return nil
	}
	switch config { //nolint:gocritic
	case "default":
		config = websocketConfDefault
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		block.AddDirectivesInSection(section, directives...)
	} else {
		block.RemoveSectionComments(section)
	}
	return nil
}

//nolint:unparam
func confSetBasicAuth(block *nginx.Block, basicAuthID string) error {
	section := "@lp_section auth_basic"
	if basicAuthID == "" {
		block.RemoveSectionComments(section)
		return nil
	}

	block.AddDirectivesInSection(section,
		nginx.NewDirective("auth_basic", []string{"\"Restricted Access\""}),
		nginx.NewDirective("auth_basic_user_file", []string{basicAuthDir + "/" + basicAuthID}),
	)
	return nil
}

func confSetProxyHeader(block *nginx.Block, config string) error {
	section := "@lp_section proxy_header"
	if config == "" {
		block.RemoveSectionComments(section)
		return nil
	}
	switch config { //nolint:gocritic
	case "default":
		config = proxyHeaderConfDefault
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		block.AddDirectivesInSection(section, directives...)
	} else {
		block.RemoveSectionComments(section)
	}
	return nil
}

func confSetLimitReq(block *nginx.Block, config string) error {
	section := "@lp_section limit_req"
	if config == "" {
		block.RemoveSectionComments(section)
		return nil
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		block.AddDirectivesInSection(section, directives...)
	} else {
		block.RemoveSectionComments(section)
	}
	return nil
}

func confSetCustom(block *nginx.Block, config string) error {
	section := "@lp_section custom"
	if config == "" {
		block.RemoveSectionComments(section)
		return nil
	}
	conf, err := nginxParseConfDefault(config)
	if err != nil {
		return apperrors.Wrap(err)
	}
	directives := conf.AllDirectives()
	if len(directives) > 0 {
		block.AddDirectivesInSection(section, directives...)
	} else {
		block.RemoveSectionComments(section)
	}
	return nil
}

func getSSLFilePaths(sslID string) (certFile, keyFile string) {
	if sslID == "" {
		return filepath.Join(certsFakeDir, "local.crt"), filepath.Join(certsFakeDir, "local.key")
	}
	return filepath.Join(certsMainDir, sslID+".crt"), filepath.Join(certsMainDir, sslID+".key")
}

func (s *nginxService) RemoveAppConfig(ctx context.Context, app *entity.App) error {
	confPath := filepath.Join(config.Current.DataPathNginxEtcConf(), app.Key+".conf")
	err := os.Remove(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperrors.Wrap(err)
	}

	// Requests nginx to reload the config files
	err = s.ReloadNginxConfig(ctx, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

//nolint:unparam
func nginxParseConfDefault(data string, options ...nginx.ParseOption) (*nginx.Config, error) {
	if data == "" {
		return nginx.NewConfig(), nil
	}
	options = append(options, func(opts *nginx.ParseOptions) {
		opts.SingleFile = true
		opts.SkipDirectiveContextCheck = true
		opts.ParseComments = true
	})
	conf, err := nginx.ParseString(data, options...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return conf, nil
}

func nginxBuildConfDefault(conf *nginx.Config, buf *bytes.Buffer, options ...nginx.BuildOption) error {
	options = append(options, func(opts *nginx.BuildOptions) {
		opts.Tabs = true
	})
	if buf.Len() > 0 {
		_, _ = buf.Write([]byte("\n\n"))
	}
	err := nginx.Build(conf, buf, options...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
