package netutil

import "strings"

func IsSubDomain(domain, sub string) bool {
	domain, _ = strings.CutPrefix(domain, "*.")
	sub, _ = strings.CutPrefix(sub, "*.")
	return strings.HasSuffix(sub, "."+domain)
}
