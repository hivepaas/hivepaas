package domainserviceimpl

import (
	"context"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) FindCertsForDomains(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	domains []string,
) (map[string]*entity.Setting, error) {
	if len(domains) == 0 {
		return nil, nil
	}
	certSettings, _, err := s.settingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(certSettings) == 0 {
		return nil, nil
	}

	timeNow := timeutil.NowUTC()
	found := make(map[string]*entity.Setting, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		for _, setting := range certSettings {
			cert, parseErr := setting.AsSSLCert()
			if parseErr != nil {
				continue
			}
			if !certCovers(cert.Domain, domain) || expired(cert, timeNow) {
				continue
			}
			if better(cert, found[domain], domain) {
				found[domain] = setting
			}
		}
	}
	return found, nil
}

// certCovers reports whether a certificate issued for certDomain can be served
// for domain.
//
// A wildcard stands for exactly one label, which is the rule TLS clients apply:
// *.example.com is the certificate for app.example.com, and neither for
// example.com itself nor for one.more.example.com.
func certCovers(certDomain, domain string) bool {
	certDomain = strings.ToLower(strings.TrimSpace(certDomain))
	if certDomain == "" {
		return false
	}
	if certDomain == domain {
		return true
	}
	parent, ok := strings.CutPrefix(certDomain, "*.")
	if !ok {
		return false
	}
	_, rest, found := strings.Cut(domain, ".")
	return found && rest == parent
}

func expired(cert *entity.SSLCert, timeNow time.Time) bool {
	return !cert.ExpireAt.IsZero() && cert.ExpireAt.Before(timeNow)
}

// better reports whether candidate should replace current as the certificate for
// domain. One issued for the name itself beats a wildcard covering it, and
// between two of the same kind the one valid for longer wins.
func better(cert *entity.SSLCert, current *entity.Setting, domain string) bool {
	if current == nil {
		return true
	}
	currentCert, err := current.AsSSLCert()
	if err != nil {
		return true
	}
	candidateExact := strings.EqualFold(cert.Domain, domain)
	currentExact := strings.EqualFold(currentCert.Domain, domain)
	if candidateExact != currentExact {
		return candidateExact
	}
	return cert.ExpireAt.After(currentCert.ExpireAt)
}
