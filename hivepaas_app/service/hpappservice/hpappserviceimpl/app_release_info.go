package hpappserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httputil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/version"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	// releaseInfoURLFormat is filled in with the branch release info is read from.
	releaseInfoURLFormat = "https://raw.githubusercontent.com/hivepaas/hivepaas/%s/release.signed.json"

	// releaseInfoBranch is where installations learn of releases. It is a branch
	// rather than a tag because it has to move with every release; what makes
	// what it serves trustworthy is the signatures, not the ref.
	releaseInfoBranch = "release"

	// releaseInfoBranchDev lets a development build see release info before it
	// is released. The signatures are required all the same.
	releaseInfoBranchDev = "main"
)

// releaseInfoURL is compiled into every binary, so an installation keeps reading
// release info from this branch until it runs a binary that says otherwise.
func releaseInfoURL() string {
	if cfg := config.Current(); cfg != nil && cfg.IsDevEnv() {
		return fmt.Sprintf(releaseInfoURLFormat, releaseInfoBranchDev)
	}
	return fmt.Sprintf(releaseInfoURLFormat, releaseInfoBranch)
}

func (s *service) GetAppReleaseInfo(ctx context.Context) (*hpappservice.AppReleaseInfo, error) {
	keys, err := loadReleaseSigningKeys(releaseKeysFS)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	fetch := func(ctx context.Context, url string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(ctx, releaseInfoFetchTimeout)
		defer cancel()
		return httputil.HTTPGet(ctx, url)
	}
	accept := func(envelope []byte) error {
		_, err := parseReleaseInfo(envelope, keys)
		return err
	}

	envelope, err := s.releaseInfoCache.get(ctx, releaseInfoURL(), time.Now(), fetch, accept)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return parseReleaseInfo(envelope, keys)
}

// parseReleaseInfo opens release.signed.json with keys given as PEM by key id,
// and decodes the release.json it carries.
//
// Nothing in release.json is looked at before the signatures verify: it names
// the images the updater will run, so an unverified file is not to be trusted
// for anything - not even for deciding that there is nothing to update.
func parseReleaseInfo(envelope []byte, keys map[string][]byte) (*hpappservice.AppReleaseInfo, error) {
	publicKeys, err := releasesig.ParsePublicKeys(keys)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := releasesig.Open(publicKeys, envelope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return decodeReleaseInfo(data)
}

// decodeReleaseInfo decodes and checks release.json. data must already have been
// verified; see parseReleaseInfo.
//
// A malformed field refuses the whole file rather than being dropped. The file is
// signed, so a malformed field is a mistake made at release - and a release that
// quietly lost its templates pin is worse to diagnose than one that fails.
func decodeReleaseInfo(data []byte) (*hpappservice.AppReleaseInfo, error) {
	info := &hpappservice.AppReleaseInfo{}
	err := json.Unmarshal(data, info)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	for channel, release := range map[string]*hpappservice.ReleaseInfo{"stable": info.Stable, "beta": info.Beta} {
		if release == nil || release.Templates == nil {
			continue
		}
		if err = validateTemplatesRef(release.Templates); err != nil {
			return nil, hperrors.Wrap(err).WithExtraDetail("%s templates", channel)
		}
	}

	if info.Stable != nil && info.Stable.AppVersion != "" {
		cmp, err := version.CmpStr(info.Stable.AppVersion, base.StableVersion.AppVersion)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		info.Stable.CanUpdate = cmp > 0
	}

	if info.Beta != nil && info.Beta.AppVersion != "" {
		cmp, err := version.CmpStr(info.Beta.AppVersion, base.BetaVersion.AppVersion)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		info.Beta.CanUpdate = cmp > 0
	}

	return info, nil
}
