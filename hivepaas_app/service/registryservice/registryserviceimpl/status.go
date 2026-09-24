package registryserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

const (
	searchQuery = `{"query":"{ RepoListWithNewestImage(requestedPage:{limit:100,offset:0,` +
		`sortBy:UPDATE_TIME}) { Results { Name Size } } }"}`

	maxSearchResponseBytes = 1 << 20
)

// Status answers the settings screen. It never fails the screen: a registry that
// cannot be reached is a status saying so, because "still restarting after a
// save" is the most common reason and it is not an error.
func (s *service) Status(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
) (*registryservice.Status, error) {
	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	status := &registryservice.Status{CredentialRotatedAt: cfg.CredentialRotatedAt, Resources: defaultResources()}
	app, err := s.loadApp(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app == nil {
		return status, nil
	}
	status.Provisioned, status.AppID = true, app.ID
	// Read off the app's service, where they live; a service that cannot be read
	// leaves the defaults rather than failing the screen.
	if res, resErr := s.systemAppService.ReadResources(ctx, app); resErr == nil && res != nil {
		status.Resources = *res
	}

	repos, stored, err := s.askRegistry(ctx, db, cfg)
	if err != nil {
		// Deliberate: a registry that cannot be reached is a status the screen
		// shows, not an error that empties it. The most common reason is that it
		// is restarting after the save that got here.
		status.Unreachable = err.Error()
		return status, nil //nolint:nilerr
	}
	status.Reachable, status.Repositories, status.StoredBytes = true, repos, stored
	return status, nil
}

// askRegistry reads what the registry holds, through its public domain and with
// the managed credential - the same path a build takes, so a status that works
// says more than a read from inside the cluster would.
func (s *service) askRegistry(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	int, int64, error) {
	user, password, err := s.credentialOf(ctx, db, cfg)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}

	url := fmt.Sprintf("https://%s/v2/_zot/ext/search", cfg.Domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(searchQuery)))
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).
			WithExtraDetail("The registry answered %d.", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSearchResponseBytes))
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	return parseSearchResponse(body)
}

// credentialOf reads the account the registry is reached with.
func (s *service) credentialOf(ctx context.Context, db database.IDB, cfg *entity.RegistrySettings) (
	string, string, error) {
	if cfg.RegistryAuthID == "" {
		return "", "", hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}

	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeRegistryAuth, cfg.RegistryAuthID, false)
	if err != nil {
		return "", "", hperrors.Wrap(err)
	}
	auth, err := setting.AsRegistryAuth()
	if err != nil {
		return "", "", hperrors.Wrap(err)
	}
	password, err := auth.Password.GetPlain()
	if err != nil {
		return "", "", hperrors.Wrap(err)
	}
	return auth.Username, password, nil
}

// parseSearchResponse reads zot's search answer. Sizes come back as strings
// because they are 64-bit, which a JSON number in a browser is not.
func parseSearchResponse(body []byte) (int, int64, error) {
	var payload struct {
		Data struct {
			RepoList struct {
				Results []struct {
					Name string `json:"Name"`
					Size string `json:"Size"`
				} `json:"Results"`
			} `json:"RepoListWithNewestImage"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	// The search extension answers a bad query with 200 and an errors array, so
	// the body is the only place a failure shows up.
	if len(payload.Errors) > 0 {
		return 0, 0, hperrors.Wrap(hperrors.ErrRegistryUnreachable).
			WithExtraDetail("%s", payload.Errors[0].Message)
	}

	var total int64
	for _, repo := range payload.Data.RepoList.Results {
		size, err := strconv.ParseInt(repo.Size, 10, 64)
		if err != nil {
			continue // a size nobody can read is not a reason to show nothing
		}
		total += size
	}
	return len(payload.Data.RepoList.Results), total, nil
}
