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
