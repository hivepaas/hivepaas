package templatemodel

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// capabilitiesPath is where a template asks for more of the host than a
// container ordinarily gets. It is the spec's own path, because a template's app
// block is a spec document.
var capabilitiesPath = []string{"deployment", "resources", "capabilities"}

// placeholderMark is enough to find a placeholder without repeating
// templaterender's grammar: every form of one starts this way.
const placeholderMark = "${{"

// Capabilities are what provisioning this template grants the app: kernel
// capabilities, sysctls, ulimits, the GPU. Nil for a template that asks for
// none, which is nearly all of them.
//
// They are read from the template rather than from a render because they are
// shown before anybody deploys: the store lists them, and the deploy dialog
// warns about them. That is also why a version may not override them - see
// validateCapabilities - so what is read here is what gets built.
func (t *Template) Capabilities() (*specmodel.Capabilities, error) {
	capabilities, problem := capabilitiesIn(t.App)
	if problem != "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", problem)
	}
	return capabilities, nil
}

// Capabilities are what this component's app is granted, and nil for one
// granted none.
func (c *Component) Capabilities() (*specmodel.Capabilities, error) {
	capabilities, problem := capabilitiesIn(c.App)
	if problem != "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", problem)
	}
	return capabilities, nil
}

// RequiresCapabilities reports whether a template asks for any. A template whose
// block cannot be read asks for something, which is what a false would hide, so
// an unreadable one counts as requiring them; validation refuses it anyway.
func (t *Template) RequiresCapabilities() bool {
	for _, app := range t.appTrees() {
		capabilities, problem := capabilitiesIn(app)
		if problem != "" || capabilities != nil {
			return true
		}
	}
	return false
}

// appTrees is every app block the template carries: its own, or one per
// component. Anything that reads what a template asks of the node has to look at
// all of them - a capability granted to one component is granted on the node.
func (t *Template) appTrees() []map[string]any {
	if !t.HasComponents() {
		return []map[string]any{t.App}
	}
	trees := make([]map[string]any, 0, len(t.Components))
	for _, component := range t.Components {
		if component != nil {
			trees = append(trees, component.App)
		}
	}
	return trees
}

// capabilitiesIn reads the block out of an app tree, and says what is wrong with
// it instead when it cannot. Both are empty for a tree that declares none.
func capabilitiesIn(app map[string]any) (*specmodel.Capabilities, string) {
	path := "app." + strings.Join(capabilitiesPath, ".")
	node := any(app)
	for _, key := range capabilitiesPath {
		parent, ok := node.(map[string]any)
		if !ok {
			return nil, ""
		}
		if node, ok = parent[key]; !ok {
			return nil, ""
		}
	}
	encoded, err := yaml.Marshal(node)
	if err != nil {
		return nil, fmt.Sprintf("%s: %s", path, err.Error())
	}
	if strings.Contains(string(encoded), placeholderMark) {
		return nil, path + ": a placeholder here would make what is granted depend on what a person fills in"
	}
	capabilities := &specmodel.Capabilities{}
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	decoder.KnownFields(true)
	if err = decoder.Decode(capabilities); err != nil {
		return nil, fmt.Sprintf("%s: %s", path, err.Error())
	}
	return capabilities, ""
}

// validateCapabilities checks the block a template declares, and refuses one
// declared anywhere else.
//
// A version may not override capabilities. What a template grants is shown in
// the store and warned about in the deploy dialog, both of which read the
// template rather than a render; a block that changed with the version chosen
// would turn that warning into a guess.
func validateCapabilities(t *Template, p *problems) {
	for i, app := range t.appTrees() {
		where := "app"
		if t.HasComponents() {
			where = fmt.Sprintf("components[%s].app", t.Components[i].Name)
		}
		capabilities, problem := capabilitiesIn(app)
		if problem != "" {
			p.add("%s", problem)
		}
		if problem = specmodel.CapabilitiesProblem(capabilities); problem != "" {
			p.add("%s.%s", where, problem)
		}
	}
	for _, version := range t.Versions {
		if version == nil || version.Override == nil {
			continue
		}
		overridden, overrideProblem := capabilitiesIn(version.Override.App)
		if overrideProblem != "" || overridden != nil {
			p.add("versions[%s].override.app.%s: capabilities are declared once, in app",
				version.Name, strings.Join(capabilitiesPath, "."))
		}
	}
}
