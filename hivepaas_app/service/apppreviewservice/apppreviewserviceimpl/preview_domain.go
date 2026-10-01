package apppreviewserviceimpl

import (
	"strings"

	"golang.org/x/net/publicsuffix"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// dnsLabelMaxLen is the longest a DNS label may be.
const dnsLabelMaxLen = 63

// previewDomain is the domain a preview answers at for one of its app's: beside
// the app's, its first label prefixed - pr-42-shop.example.com for
// shop.example.com - rather than under it, as pr-42.shop.example.com would be.
//
// A wildcard covers one label, so beside is where the app's own are already
// covered: the certificate of *.example.com the app has, the DNS record of
// *.example.com that points at the servers. Under the app's domain, every
// preview needed a record and a certificate of its own.
//
// The app's domain at the apex of its zone, example.com, has no label to
// prefix that would stay in the zone: its previews are under it,
// pr-42.example.com, which *.example.com covers too. So is a subdomain asked
// for with dots in it, as given.
func previewDomain(subdomain, domain string) (string, error) {
	domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
	subdomain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(subdomain), "."+domain))

	first, rest, nested := strings.Cut(domain, ".")
	if !nested || isZoneApex(domain) || strings.Contains(subdomain, ".") {
		return subdomain + "." + domain, nil
	}

	label := subdomain + "-" + first
	if len(label) > dnsLabelMaxLen {
		return "", hperrors.NewArgumentInvalid("preview subdomain").WithMsgLog(
			"the preview's domain label %s is longer than %d characters: choose a shorter subdomain",
			label, dnsLabelMaxLen)
	}
	return label + "." + rest, nil
}

// isZoneApex reports whether a domain is the registrable part of its name, or
// one the public suffix list cannot place - such as a bare suffix - which is
// treated alike: nothing beside it would stay in its zone.
func isZoneApex(domain string) bool {
	apex, err := publicsuffix.EffectiveTLDPlusOne(domain)
	return err != nil || apex == domain
}
