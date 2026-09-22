package templatemodel

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// MaxComponents caps the apps one template creates for itself. Appwrite is
// twenty-nine services and Supabase thirteen, which is where this number comes
// from: it is a guard against a template that should have stayed a Compose file,
// not a target to fill.
//
// What keeps a stack from becoming a fan-out is the rule beside it: components
// share the template's dependencies rather than declaring their own, so a
// request creates this list plus the dependencies, and nothing further.
const MaxComponents = 30

// componentNamePattern is the dependency name pattern: the name becomes the
// suffix of an app name, and that app's key is what environment references use,
// so it has to come through slugifying unchanged and leave room for the app name
// in front of it.
var componentNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]{0,15}$`)

// Component is one app of a template that creates several: the gateway, the
// authentication service, one of a dozen workers sharing an image.
//
// It is declared inline rather than naming another template, because the
// services of one application are not reusable on their own - twenty-three of
// Appwrite's are the same image with a different command, and as separate
// templates they would be twenty-three files differing by a line.
type Component struct {
	// Name is the role, and the suffix of the created app's name.
	Name  string `yaml:"name"`
	Title string `yaml:"title"`
	// Primary marks the one component that owns the template's domain and is the
	// app the others are created for. Exactly one component has it.
	Primary bool `yaml:"primary,omitempty"`
	// Needs names components this one is created and started after.
	Needs []string `yaml:"needs,omitempty"`
	// App is a specmodel.AppDoc with placeholders in it, as the template's own
	// app block is for a template with one app.
	App map[string]any `yaml:"app"`
}

// VersionComponent is what a version pins for one component. It is a struct
// rather than a bare string so that a version can later say more about a
// component than which image it runs.
type VersionComponent struct {
	Image string `yaml:"image"`
}

// HasComponents reports whether the template creates several apps of its own.
func (t *Template) HasComponents() bool {
	return len(t.Components) > 0
}

func (t *Template) FindComponent(name string) *Component {
	for _, component := range t.Components {
		if component != nil && component.Name == name {
			return component
		}
	}
	return nil
}

// PrimaryComponent is the component that owns the domain, nil for a template
// without components. Validation guarantees exactly one, so a nil here on a
// template with components is a template that never passed it.
func (t *Template) PrimaryComponent() *Component {
	for _, component := range t.Components {
		if component != nil && component.Primary {
			return component
		}
	}
	return nil
}

// ComponentImage is this version's image for one component.
func (v *Version) ComponentImage(component string) string {
	entry := v.Components[component]
	if entry == nil {
		return ""
	}
	return entry.Image
}

// ComponentAppName is the name of the app created for a component. It has the
// same shape as a dependency's, because both are apps created beside the one the
// person named.
func ComponentAppName(appName, componentName string) string {
	return appName + "-" + componentName
}

// ComponentOrder sorts the components so that every one comes after those it
// needs. Components that need nothing keep their declared order, which is what
// makes a template's file order the order a person sees.
//
// It reports an error on a cycle rather than looping. Validation refuses cycles
// first, so reaching that here means a template that never passed it.
func (t *Template) ComponentOrder() ([]*Component, error) {
	ordered := make([]*Component, 0, len(t.Components))
	placed := make(map[string]bool, len(t.Components))
	for len(ordered) < len(t.Components) {
		progressed := false
		for _, component := range t.Components {
			if component == nil || placed[component.Name] {
				continue
			}
			if !slices.ContainsFunc(component.Needs, func(need string) bool { return !placed[need] }) {
				ordered = append(ordered, component)
				placed[component.Name] = true
				progressed = true
			}
		}
		if !progressed {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: components: needs form a cycle", t.Metadata.Name)
		}
	}
	return ordered, nil
}

// validateComponents checks what a template says about the apps it creates for
// itself, and the exclusivity between components and the single app block.
func validateComponents(t *Template, p *problems) {
	if !t.HasComponents() {
		if len(t.App) == 0 {
			p.add("app is required")
		}
		return
	}
	if len(t.App) > 0 {
		p.add("components and app are exclusive: a template has one or the other")
	}
	if len(t.Variants) > 0 {
		// A variant multiplies the image table by the component table, and nothing
		// wanted needs both yet. See the components design, §2.
		p.add("components and variants cannot be used together")
	}
	if len(t.Components) > MaxComponents {
		p.add("components: at most %d", MaxComponents)
	}

	declared := make(map[string]bool, len(t.Components))
	primaries := 0
	for i, component := range t.Components {
		if component == nil {
			p.add("components[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("components[%s]", component.Name)
		if !componentNamePattern.MatchString(component.Name) {
			p.add("components[%d].name %q must be lowercase letters and digits, starting with a letter", i,
				component.Name)
		}
		if declared[component.Name] {
			p.add("%s is declared twice", prefix)
		}
		declared[component.Name] = true
		if component.Title == "" {
			p.add("%s.title is required", prefix)
		}
		if len(component.App) == 0 {
			p.add("%s.app is required", prefix)
		}
		if component.Primary {
			primaries++
		} else {
			validateSecondaryRouting(prefix, component, p)
		}
	}
	if primaries != 1 {
		p.add("exactly one component must be primary")
	}
	validateComponentNeeds(t, declared, p)
}

// validateSecondaryRouting refuses a domain on a component that is not the
// primary. An application has one front door: the others are reached inside the
// project by key, and a template that put a domain on two of them would be
// asking for one address to route to two apps.
//
// A port is allowed, because a component that speaks HTTP to its siblings still
// has one.
func validateSecondaryRouting(prefix string, component *Component, p *problems) {
	settings, _ := component.App["settings"].(map[string]any)
	routing, _ := settings[specmodel.SingletonBlockName(base.SettingTypeAppRouting)].(map[string]any)
	if routing == nil {
		return
	}
	if exposed, _ := routing["exposePublicly"].(bool); exposed {
		p.add("%s: only the primary component may be exposed publicly", prefix)
	}
	if domains, found := routing["domains"]; found && domains != nil {
		p.add("%s: only the primary component may have domains", prefix)
	}
}

// validateComponentNeeds checks that every need names a declared component and
// that the graph has no cycle. A cycle is found by ordering, which is the same
// walk provisioning does.
func validateComponentNeeds(t *Template, declared map[string]bool, p *problems) {
	sound := true
	for _, component := range t.Components {
		if component == nil {
			continue
		}
		prefix := fmt.Sprintf("components[%s]", component.Name)
		seen := map[string]bool{}
		for _, need := range component.Needs {
			switch {
			case need == component.Name:
				p.add("%s.needs: a component cannot need itself", prefix)
				sound = false
			case !declared[need]:
				p.add("%s.needs: %q is not a declared component", prefix, need)
				sound = false
			case seen[need]:
				p.add("%s.needs: %q is named twice", prefix, need)
			}
			seen[need] = true
		}
	}
	if !sound {
		// Ordering a graph with a need that names nothing would report a cycle that
		// is not there, on top of the problem already reported.
		return
	}
	if _, err := t.ComponentOrder(); err != nil {
		p.add("components: needs form a cycle")
	}
}
