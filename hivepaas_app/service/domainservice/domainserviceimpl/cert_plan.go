package domainserviceimpl

import (
	"net"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// certPlan is one certificate to obtain, and the domains waiting on it.
type certPlan struct {
	// Name is what the certificate is issued for: the domain itself, or the
	// wildcard covering it.
	Name string
	// Wildcard decides where it is kept. A wildcard covers every app under a
	// parent domain, including apps in other projects and apps that do not exist
	// yet, so it belongs to the installation rather than to the project that
	// happened to ask for it first. A name issued for one app does not.
	Wildcard bool
	Domains  []string
}

// planCerts decides what to obtain for the domains nothing already covers.
//
// A DNS provider turns the answer into a wildcard: it covers every other app
// under the same parent, it can be issued before the domain points anywhere, and
// it spends one of a certificate authority's weekly allowances instead of one per
// app. Without a DNS provider the only challenge left is HTTP-01, which reaches
// the app through its own address, so the certificate is for that name alone.
//
// publicAuthority is what makes a local name hopeless: an authority validates
// over the public internet, and app.localhost is not there. An installation
// signing its own certificates has no such limit and gets one for every name.
//
// skipped says why a domain was left without one, in words a person can act on.
func planCerts(domains []string, hasDNSProvider, publicAuthority bool) (plans []*certPlan, skipped map[string]string) {
	skipped = map[string]string{}
	byName := map[string]*certPlan{}
	var order []string

	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if publicAuthority {
			if reason := unobtainable(domain); reason != "" {
				skipped[domain] = reason
				continue
			}
		}

		name, wildcard := domain, false
		if hasDNSProvider {
			if parent := wildcardFor(domain); parent != "" {
				name, wildcard = parent, true
			}
		}
		plan := byName[name]
		if plan == nil {
			plan = &certPlan{Name: name, Wildcard: wildcard}
			byName[name] = plan
			order = append(order, name)
		}
		plan.Domains = append(plan.Domains, domain)
	}

	for _, name := range order {
		plans = append(plans, byName[name])
	}
	return plans, skipped
}

// wildcardFor is the wildcard that covers a domain: one label replaced, because
// that is all a wildcard matches. It is empty for a name with nothing to the
// left of the registrable part, where the wildcard would have to be issued for
// the zone apex and cover far more than the app asking.
func wildcardFor(domain string) string {
	labels := strings.Split(domain, ".")
	if len(labels) < 3 { //nolint:mnd // example.com has nothing to wildcard
		return ""
	}
	return "*." + strings.Join(labels[1:], ".")
}

// unobtainable reports why no certificate authority would issue for a name, so
// the attempt is not made at all: a public authority validates over the public
// internet, and these names never resolve there.
func unobtainable(domain string) string {
	if net.ParseIP(domain) != nil {
		return "an address rather than a name"
	}
	if !strings.Contains(domain, ".") {
		return "a name with no domain part"
	}
	for _, suffix := range privateSuffixes {
		if strings.HasSuffix(domain, suffix) {
			return "a name reserved for local use (" + suffix + ")"
		}
	}
	return ""
}

// privateSuffixes are the names reserved for local and testing use - RFC 6761
// and RFC 8375 - which no public certificate authority issues for.
var privateSuffixes = []string{
	".localhost", ".local", ".internal", ".home.arpa",
	".test", ".example", ".invalid",
}

// certTypeOf is the kind of certificate a policy asks for. Let's Encrypt is what
// an installation that has never chosen gets: it is free, it needs no account,
// and it is what a public name is served with.
func certTypeOf(certSettings *entity.DomainCertSettings) base.SSLCertType {
	if certSettings == nil || certSettings.CertType == "" {
		return base.SSLCertTypeLetsEncrypt
	}
	return certSettings.CertType
}

// certSettingFor builds the setting a plan becomes: where it lives, what it is
// called, and what obtaining it will use.
func certSettingFor(
	plan *certPlan,
	certSettings *entity.DomainCertSettings,
	projectID string,
	provider, acmeProvider entity.ObjectID,
	settingID string,
	timeNow time.Time,
) (*entity.Setting, *entity.SSLCert) {
	scope, objectID := base.ObjectScopeProject, projectID
	if plan.Wildcard {
		// Global: a wildcard covers every project under the parent domain,
		// including projects that do not exist yet, and a second copy of it would
		// only spend another issuance.
		scope, objectID = base.ObjectScopeGlobal, ""
	}

	certType := certTypeOf(certSettings)
	keyType := certSettings.KeyType
	if keyType == "" {
		keyType = base.SSLKeyTypeDefault
	}

	setting := &entity.Setting{
		ID:       settingID,
		Scope:    scope,
		ObjectID: objectID,
		Type:     base.SettingTypeSSLCert,
		Kind:     string(certType),
		Status:   base.SettingStatusActive,
		Name:     plan.Name,
		// Inheritable, always: a certificate is kept at the scope that owns it and
		// used by the apps under it. One an app could not see would be a
		// certificate nothing can be served with.
		Inheritable: true,
		Version:     entity.CurrentSSLCertVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	cert := &entity.SSLCert{
		CertType: certType,
		Domain:   plan.Name,
		KeyType:  keyType,
		Email:    certSettings.Email,
		// ValidPeriod is how long a certificate the installation signs itself is
		// good for. An authority decides the lifetime of the ones it issues and
		// ignores this, so it is carried for both and used by one.
		ValidPeriod:  certSettings.ValidPeriod,
		AutoRenew:    true,
		Provider:     provider,
		AcmeProvider: acmeProvider,
	}
	return setting, cert
}
