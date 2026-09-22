package entity

import (
	"maps"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	CurrentAppTemplateSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppTemplate, &appTemplateSettingsParser{})

type appTemplateSettingsParser struct {
}

func (s *appTemplateSettingsParser) New() SettingData {
	return &AppTemplateSettings{}
}

// AppTemplateSettings records the template an app was provisioned from.
//
// Phase 1 writes it and reads it only for display. It is complete now so that
// phase 2 can update the apps phase 1 created without a migration that would
// have to reconstruct a base nobody recorded.
//
// TODO: app templates phase 2 - UpdatePolicy ("follow" or "pinned"). See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type AppTemplateSettings struct {
	Source   string `json:"source"`
	Template string `json:"template"`
	Title    string `json:"title"`
	// Version is the major line, such as "17".
	Version string `json:"version"`
	Variant string `json:"variant,omitempty"`
	// ImageOverride is the image the user chose instead of the template's, empty
	// when the template's own image is in use. It is stored so phase 2's update can
	// say "you chose this" rather than guessing from a string comparison.
	ImageOverride string `json:"imageOverride,omitempty"`

	// Dependencies are the apps created alongside this one because its template
	// named them, in the order they were created.
	Dependencies []AppTemplateDependency `json:"dependencies,omitempty"`
	// Component is the role this app plays in a template that creates several -
	// "auth", "gw" - empty for an app that is its template's only one. The
	// primary component carries its own name here as well, because "which part of
	// the stack is this" has to be answerable on every app of it.
	Component string `json:"component,omitempty"`
	// Components are the other apps created for this one because its template
	// declared them, in the order they were created. Set on the primary app only,
	// as Dependencies is.
	Components []AppTemplateComponent `json:"components,omitempty"`
	// CreatedForAppID is the app this one was created to serve, empty when it was
	// created on its own. This app is that app's logical child, so deleting that
	// app deletes this one with it - see appservice.DeleteApp's cascade over
	// LogicalParentID.
	CreatedForAppID string `json:"createdForAppId,omitempty"`

	// Params are the values given at creation, secrets encrypted. Parameters are
	// fixed after creation: a change is made on the app, and phase 2's merge sees
	// it as a user change.
	Params map[string]*AppTemplateParam `json:"params,omitempty"`

	Base AppTemplateBase `json:"base"`
}

type AppTemplateParam struct {
	Value  string         `json:"value,omitempty"`
	Secret EncryptedField `json:"secret,omitzero"`
}

type AppTemplateDependency struct {
	// Name is the role the template gave it: db, cache.
	Name     string `json:"name"`
	AppID    string `json:"appId"`
	Template string `json:"template"`
}

// AppTemplateComponent is one of the other apps a template that creates several
// made for this one. It carries no template name because a component is declared
// inside the template rather than naming another.
type AppTemplateComponent struct {
	// Name is the role the template gave it: auth, worker.
	Name  string `json:"name"`
	AppID string `json:"appId"`
}

// AppTemplateBase is what the template rendered to when it was last applied:
// the base of phase 2's three-way merge.
//
// It is the render, not the template file. A later HivePaaS that fixes a bug in
// rendering would render the old template differently from what was applied,
// and every app would show changes no template made. Rendered holds secret
// placeholders, never secrets.
type AppTemplateBase struct {
	Revision       string    `json:"revision"`
	Release        string    `json:"release"`
	TemplateSHA256 string    `json:"templateSha256"`
	Rendered       string    `json:"rendered"`
	RenderedSHA256 string    `json:"renderedSha256"`
	AppliedAt      time.Time `json:"appliedAt"`
}

func (s *AppTemplateSettings) GetType() base.SettingType {
	return base.SettingTypeAppTemplate
}

// GetRefObjectIDs is empty on purpose. A volume parameter records the id the
// user chose; the reference itself lives in the app's mount, where the storage
// settings already track it.
func (s *AppTemplateSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppTemplateSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt implements the secret decrypter the settings reveal path looks for.
func (s *AppTemplateSettings) Decrypt() error {
	for _, name := range slices.Sorted(maps.Keys(s.Params)) {
		param := s.Params[name]
		if param == nil || param.Secret.IsEmpty() {
			continue
		}
		if _, err := param.Secret.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Setting) AsAppTemplateSettings() (*AppTemplateSettings, error) {
	return parseSettingAs[*AppTemplateSettings](s)
}

func (s *Setting) MustAsAppTemplateSettings() *AppTemplateSettings {
	return gofn.Must(s.AsAppTemplateSettings())
}
