package githubappuc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

const redirectPage = `<html>
	<body>
		<h3>Redirecting...</h3>
		<form id="new-app-form" action="{{.Action}}" method="post">
			<input type="hidden" name="manifest" id="manifest" value={{.Manifest}}>
		</form>
		<script>document.getElementById("new-app-form").submit()</script>
	</body>
</html>`

type redirectTemplate struct {
	Manifest string
	Action   string
}

func (uc *UC) BeginGithubAppManifestFlowCreation(
	ctx context.Context,
	req *githubappdto.BeginGithubAppManifestFlowCreationReq,
) (*githubappdto.BeginGithubAppManifestFlowCreationResp, error) {
	manifestCache, err := uc.cacheAppManifestRepo.Get(ctx, req.SettingID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	manifestJSON, err := json.Marshal(manifestCache.Manifest)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	githubApp := manifestCache.GithubApp.MustAsGithubApp()

	actionURL := fmt.Sprintf("%s/new?state=%v",
		entity.GithubAppOwnerSettingsBaseURL(githubApp.Organization), req.State)

	data := &redirectTemplate{
		Action:   actionURL,
		Manifest: string(manifestJSON),
	}

	buf := bytes.NewBuffer(make([]byte, 10000)) //nolint:mnd
	tmpl := template.Must(template.New("redirect").Parse(redirectPage))
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &githubappdto.BeginGithubAppManifestFlowCreationResp{
		Data: &githubappdto.BeginGithubAppManifestFlowCreationDataResp{
			PageContent: buf.String(),
		},
	}, nil
}
