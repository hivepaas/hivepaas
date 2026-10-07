package apptemplateuc

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// indexTemplateService answers from an index, and refuses every template the way
// the service refuses one written for a newer HivePaaS - or with err when set.
type indexTemplateService struct {
	apptemplateservice.Service
	index *templatemodel.Index
	err   error
}

func (f *indexTemplateService) Index(context.Context) (*apptemplateservice.IndexResp, error) {
	return &apptemplateservice.IndexResp{Source: "official", Revision: "rev", Index: f.index}, nil
}

func (f *indexTemplateService) Template(context.Context, string) (*apptemplateservice.TemplateResp, error) {
	if f.err != nil {
		return nil, f.err
	}
	return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible)
}

func newerEntry(versionCode string) *templatemodel.IndexEntry {
	return &templatemodel.IndexEntry{
		Name: "newer", Title: "Newer", Tagline: "Written for a newer HivePaaS",
		Categories: []string{"webapps/dev-tools"},
		File:       templatemodel.FileRef{Path: "templates/newer.yaml", SHA256: "aa"},
		Icon:       templatemodel.FileRef{Path: "icons/newer.svg", SHA256: "bbbbbbbbbbbb"},
		Components: []*templatemodel.IndexComponent{{Name: "web", Title: "Web", Primary: true}},
		Versions:   []*templatemodel.IndexVersion{{Name: "2", Release: "2.0", Default: true}},
		Requires:   templatemodel.Requires{VersionCode: versionCode},
	}
}

// A template written for a newer HivePaaS still opens in the store, from what the
// index says of it: no form, and marked as needing a newer HivePaaS - not an
// error page for something the listing showed a moment ago.
func TestGetAppTemplateShowsANewerTemplateFromTheIndex(t *testing.T) {
	previous := config.Current()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(previous) })
	uc := &UC{appTemplateService: &indexTemplateService{
		index: &templatemodel.Index{Templates: []*templatemodel.IndexEntry{newerEntry("v999999")}},
	}}

	resp, err := uc.GetAppTemplate(context.Background(), nil, &apptemplatedto.GetAppTemplateReq{Name: "newer"})

	assert.NoError(t, err)
	tmpl := resp.Data
	assert.False(t, tmpl.Compatible)
	assert.Equal(t, "Newer", tmpl.Title)
	assert.Equal(t, "official", tmpl.Source)
	assert.Equal(t, "2", tmpl.Versions[0].Name)
	assert.Equal(t, "web", tmpl.Components[0].Name)

	// Every list is a list, so the dashboard reads lengths without guarding them;
	// only the objects a template may lack are null.
	data, err := json.Marshal(tmpl)
	assert.NoError(t, err)
	var decoded any
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.Empty(t, nullsBesides(decoded, "", "links", "capabilities", "dockerApi"))
	assert.Contains(t, string(data), `"parameters":[]`)
	assert.Contains(t, string(data), `"tags":[]`)
}

// nullsBesides lists the keys holding null anywhere in a decoded JSON value,
// except those named.
func nullsBesides(value any, key string, allowed ...string) []string {
	switch value := value.(type) {
	case nil:
		if slices.Contains(allowed, key) {
			return nil
		}
		return []string{key}
	case map[string]any:
		var out []string
		for k, v := range value {
			out = append(out, nullsBesides(v, k, allowed...)...)
		}
		return out
	case []any:
		var out []string
		for _, v := range value {
			out = append(out, nullsBesides(v, key, allowed...)...)
		}
		return out
	}
	return nil
}

// Refused for another reason - a dependency needing a newer HivePaaS, while the
// template itself does not - the refusal stands: the template cannot be offered,
// and its own entry would say it can.
func TestGetAppTemplateKeepsARefusalTheIndexDoesNotExplain(t *testing.T) {
	uc := &UC{appTemplateService: &indexTemplateService{
		index: &templatemodel.Index{Templates: []*templatemodel.IndexEntry{newerEntry("v000001")}},
	}}

	_, err := uc.GetAppTemplate(context.Background(), nil, &apptemplatedto.GetAppTemplateReq{Name: "newer"})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateIncompatible)
}

func TestGetAppTemplatePassesOtherErrorsOn(t *testing.T) {
	uc := &UC{appTemplateService: &indexTemplateService{
		index: &templatemodel.Index{},
		err:   hperrors.Wrap(hperrors.ErrAppTemplateInvalid),
	}}

	_, err := uc.GetAppTemplate(context.Background(), nil, &apptemplatedto.GetAppTemplateReq{Name: "broken"})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}
