package netutil

import "strings"

func IsSubDomain(domain, sub string) bool {
	domain, _ = strings.CutPrefix(domain, "*.")
	sub, _ = strings.CutPrefix(sub, "*.")
	return strings.HasSuffix(sub, "."+domain)
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
