package domainhelper

import "strings"

func IsSubdomain(domain, sub string) bool {
	domain, _ = strings.CutPrefix(domain, "*.")
	sub, _ = strings.CutPrefix(sub, "*.")
	return strings.HasSuffix(sub, "."+domain)
}

func IsSubdomainOrEqual(domain, sub string) bool {
	domain, _ = strings.CutPrefix(domain, "*.")
	sub, _ = strings.CutPrefix(sub, "*.")
	return domain == sub || strings.HasSuffix(sub, "."+domain)
}

func CalcMatchingDomains(subdomain string) (res []string) {
	res = append(res, subdomain)
	domain := strings.Trim(subdomain, "*.")
	for {
		var found bool
		_, domain, found = strings.Cut(domain, ".")
		if !found {
			break
		}
		res = append(res, domain, "*."+domain)
	}
	return res
}

func IsDomainAllowed(domain string, allowedList []string) bool {
	for _, allowed := range allowedList {
		if IsSubdomainOrEqual(allowed, domain) {
			return true
		}
	}
	return false
}

// IsDomainCoveredByCert reports whether a certificate issued for certDomain can
// be served for domain.
//
// A wildcard stands for exactly one label, which is the rule TLS clients apply:
// *.example.com is the certificate for app.example.com, and neither for
// example.com itself nor for one.more.example.com.
func IsDomainCoveredByCert(domain, certDomain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))
	certDomain = strings.ToLower(strings.TrimSpace(certDomain))
	if domain == "" || certDomain == "" {
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
