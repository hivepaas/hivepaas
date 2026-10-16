package config

import (
	"errors"
	"fmt"
	"net/url"
)

// ErrLicensePortalURLInvalid is a portal URL that is not an http or https URL
// with a host: every link to the portal and every refresh would be wrong.
var ErrLicensePortalURLInvalid = errors.New("license portal URL is invalid")

// DefaultLicensePortalURL is the portal production installations buy keys from
// and ask for newer ones.
const DefaultLicensePortalURL = "https://portal.hivepaas.com"

// License is how this installation reaches the license portal. Online refresh
// sends the portal the stored key and this installation's ID, and nothing else.
// See docs/superpowers/specs/2026-10-16-licensing-design.md, "Online renewal".
//
// The worker runs the daily refresh and does not get the app's config reload: a
// change takes effect there on its next start.
type License struct {
	// PortalURL is the portal's web app, with its API under /api. Outside
	// development it must be https: the portal's answer is not signed as a
	// whole, so a "moved" must not cross the network in the clear.
	PortalURL string `toml:"portal_url" env:"HP_LICENSE_PORTAL_URL" default:"https://portal.hivepaas.com"`
	// RefreshDisabled turns off the daily check for a newer key. It is negative on
	// purpose: configor cannot give a bool a default of true that a config can
	// turn off. A person asking now, from System > License, is still answered.
	RefreshDisabled bool `toml:"refresh_disabled" env:"HP_LICENSE_REFRESH_DISABLED"`
}

// Validate refuses a portal URL that is not an absolute http or https URL. Its
// scheme is checked against the environment when a refresh is made, as only a
// refresh sends anything.
func (l License) Validate() error {
	u, err := url.Parse(l.PortalURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("%w: HP_LICENSE_PORTAL_URL must be an http or https URL such as %s, not %q",
			ErrLicensePortalURLInvalid, DefaultLicensePortalURL, l.PortalURL)
	}
	return nil
}
