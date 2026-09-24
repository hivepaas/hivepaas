package specmodel

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// What an app document asks of the installation it lands on, read from the
// document before anything is built. Template creation and import both check
// these, and each asks its own services: creation refuses on the first finding,
// import reports every one and leaves out the objects it rewrites.

// ExternalRefKey is the key an external reference is written under, in place of
// the identifier of a setting the export did not hold.
const ExternalRefKey = "external"

// PublishedPorts are the ports a document publishes on every node of the
// cluster.
func PublishedPorts(doc *AppDoc) []PortConfig {
	if doc == nil || doc.Deployment == nil || doc.Deployment.Networks == nil ||
		doc.Deployment.Networks.EndpointSpec == nil {
		return nil
	}
	ports := make([]PortConfig, 0, len(doc.Deployment.Networks.EndpointSpec.Ports))
	for _, port := range doc.Deployment.Networks.EndpointSpec.Ports {
		if port != nil && port.Published > 0 {
			ports = append(ports, *port)
		}
	}
	return ports
}

// ActiveDomains are the addresses a document's routing answers at. A domain's
// certificate is not one of them, so a certificate named by an external
// reference is read as none.
func ActiveDomains(doc *AppDoc) ([]string, error) {
	if doc == nil {
		return nil, nil
	}
	body, ok := doc.Settings[SingletonBlockName(base.SettingTypeAppRouting)]
	if !ok {
		return nil, nil
	}
	encoded, err := json.Marshal(withoutExternalRefs(body))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	routing := &entity.AppRoutingSettings{}
	if err = json.Unmarshal(encoded, routing); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return routing.GetActiveDomainNames(), nil
}

// withoutExternalRefs copies a body with every external reference replaced by
// an empty identifier.
func withoutExternalRefs(node any) any {
	switch typed := node.(type) {
	case map[string]any:
		if _, external := typed[ExternalRefKey]; external && len(typed) == 1 {
			return ""
		}
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = withoutExternalRefs(value)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = withoutExternalRefs(value)
		}
		return out
	}
	return node
}

// GrantedCapabilities names, for a person to read, what a document asks the
// host for beyond what a container ordinarily gets. It is empty for a document
// that carries no capabilities.
func GrantedCapabilities(doc *AppDoc) []string {
	if doc == nil || doc.Deployment == nil || doc.Deployment.Resources == nil {
		return nil
	}
	capabilities := doc.Deployment.Resources.Capabilities
	if capabilities == nil {
		return nil
	}
	granted := make([]string, 0, len(capabilities.CapabilityAdd))
	granted = append(granted, capabilities.CapabilityAdd...)
	if capabilities.EnableGPU {
		granted = append(granted, docker.CapabilityGPU)
	}
	for _, described := range []struct {
		what  string
		count int
	}{
		{"sysctls", len(capabilities.Sysctls)}, {"ulimits", len(capabilities.Ulimits)},
	} {
		if described.count > 0 {
			granted = append(granted, described.what)
		}
	}
	if len(granted) == 0 && len(capabilities.CapabilityDrop) == 0 && capabilities.OomScoreAdj == 0 {
		return nil
	}
	if len(granted) == 0 {
		// A block that only drops capabilities or nudges the OOM score grants
		// nothing, but it is still the privileged block, and the gate is the
		// block's rather than each field's.
		granted = append(granted, string(BlockDeploymentResources)+".capabilities")
	}
	return granted
}
