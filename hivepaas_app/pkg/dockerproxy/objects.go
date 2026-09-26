package dockerproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const (
	kindVolumes  = "volumes"
	kindNetworks = "networks"

	fieldName   = "Name"
	fieldDriver = "Driver"
	// local is docker's name for its own volume driver, log driver and network
	// scope: what stays on this node.
	local = "local"

	// serviceIDLabel is how Docker marks the containers of a swarm service's tasks.
	serviceIDLabel = "com.docker.swarm.service.id"
)

var (
	volumeCreateFields  = []string{fieldName, fieldDriver, fieldLabels}
	volumeDrivers       = []string{"", local}
	networkCreateFields = []string{fieldName, "CheckDuplicate", fieldDriver, "Scope", "Internal", "Attachable",
		"EnableIPv4", "EnableIPv6", fieldLabels, "IPAM"}
	networkDrivers       = []string{"", "bridge"}
	networkScopes        = []string{"", local}
	ipamDrivers          = []string{"", "default"}
	networkConnectFields = []string{"Container", "EndpointConfig", "Force"}
	// endpointFields are what a child may say about a network it joins. A static
	// address or a driver option is a decision about the network, not the child.
	endpointFields = []string{"Aliases", "DNSNames"}
	// reservedLabelPrefixes are labels a client may not set: Docker's, which mark
	// swarm tasks, and HivePaaS's, which mark what it owns.
	reservedLabelPrefixes = []string{"com.docker.", "hivepaas."}
)

// labeled is the part of an inspect answer that says what an object is called
// and whose it is.
type labeled struct {
	Name   string
	Labels map[string]string
}

func (p *Proxy) inspect(ctx context.Context, kind, id string) (labeled, bool, error) {
	var info labeled
	found, err := p.daemon.get(ctx, "/"+kind+"/"+url.PathEscape(id), &info)
	return info, found, err
}

// ownLabels returns the client's labels less the reserved ones, with the owner.
func ownLabels(labels map[string]any, appID string) map[string]any {
	out := map[string]any{}
	for key, value := range labels {
		reserved := slices.ContainsFunc(reservedLabelPrefixes, func(prefix string) bool {
			return strings.HasPrefix(key, prefix)
		})
		if !reserved {
			out[key] = value
		}
	}
	out[OwnerLabel] = appID
	return out
}

func (p *Proxy) listVolumes(c *call) {
	var answer map[string]any
	if !p.fetch(c, &answer) {
		return
	}
	items := make([]map[string]any, 0, len(list(answer["Volumes"])))
	for _, item := range list(answer["Volumes"]) {
		items = append(items, object(item))
	}
	kept := keep(items, func(item map[string]any) bool { return ownedBy(item[fieldLabels], c.policy.AppID) })
	answer["Volumes"] = kept
	p.decide(c, true, fmt.Sprintf("%d of %d volumes are the app's", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, answer)
}

func (p *Proxy) volumeCreate(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = checkFields("Volume", body, volumeCreateFields, nil)
	}
	if _, shared := c.policy.sharedVolume(text(body[fieldName])); err == nil && shared {
		// Creating a volume that exists answers with it; a client that creates
		// before it mounts must not be handed a new, empty one of the same name.
		p.decide(c, true, "a shared volume of the app, which exists")
		writeJSON(c.w, http.StatusCreated, sharedVolumeInfo(c.policy, text(body[fieldName])))
		return
	}
	if err == nil {
		err = refuseReserved(c.policy, text(body[fieldName]))
	}
	if driver := text(body[fieldDriver]); err == nil && !slices.Contains(volumeDrivers, driver) {
		err = refusef("volume driver %s is not allowed", driver)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), c.policy.AppID)
	setBody(c.r, body)
	p.forward(c, "a volume of the app's own")
}

func (p *Proxy) onVolume(c *call) {
	if _, shared := c.policy.sharedVolume(c.args[0]); shared {
		if c.r.Method != http.MethodGet {
			p.refuse(c, refusef("volume %s is a directory of the app, and is not removed through Docker", c.args[0]))
			return
		}
		p.decide(c, true, "a shared volume of the app")
		writeJSON(c.w, http.StatusOK, sharedVolumeInfo(c.policy, c.args[0]))
		return
	}
	info, found, err := p.inspect(c.r.Context(), kindVolumes, c.args[0])
	if err == nil && (!found || info.Labels[OwnerLabel] != c.policy.AppID) {
		err = refusef("volume %s is not one this app created", c.args[0])
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a volume of the app's own")
}

// sharedVolumeInfo is what the proxy answers for a shared volume in Docker's
// words. There is no volume of that name to inspect: the name stands for a
// directory of one the app mounts, and where that is on the node is not the
// child's to know.
func sharedVolumeInfo(policy *Policy, name string) map[string]any {
	return map[string]any{
		fieldName: name, fieldDriver: local, "Scope": local, "Mountpoint": "", "Options": map[string]any{},
		fieldLabels: map[string]any{OwnerLabel: policy.AppID},
	}
}

func (p *Proxy) listNetworks(c *call) {
	var items []map[string]any
	if !p.fetch(c, &items) {
		return
	}
	kept := keep(items, func(item map[string]any) bool {
		return ownedBy(item[fieldLabels], c.policy.AppID) || c.policy.joinable(text(item[fieldName]))
	})
	p.decide(c, true, fmt.Sprintf("%d of %d networks are the app's", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, kept)
}

func (p *Proxy) networkCreate(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = checkNetworkCreate(body)
	}
	if err == nil {
		err = refuseReserved(c.policy, text(body[fieldName]))
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), c.policy.AppID)
	setBody(c.r, body)
	p.forward(c, "a network of the app's own")
}

// refuseReserved refuses a name HivePaaS keeps for the objects it makes for
// apps.
func refuseReserved(policy *Policy, name string) error {
	if policy.reserved(name) {
		return refusef("%s is a name HivePaaS uses for an app's own socket and network", name)
	}
	return nil
}

// checkNetworkCreate takes a plain bridge on this node. Anything else - an
// overlay across the cluster, a driver option naming a host interface, an
// address range of the operator's - is a decision about the host.
func checkNetworkCreate(body map[string]any) error {
	if err := checkFields("Network", body, networkCreateFields, nil); err != nil {
		return err
	}
	if driver := text(body[fieldDriver]); !slices.Contains(networkDrivers, driver) {
		return refusef("network driver %s is not allowed", driver)
	}
	if scope := text(body["Scope"]); !slices.Contains(networkScopes, scope) {
		return refusef("network scope %s is not allowed", scope)
	}
	ipam := object(body["IPAM"])
	if driver := text(ipam[fieldDriver]); !slices.Contains(ipamDrivers, driver) {
		return refusef("IPAM driver %s is not allowed", driver)
	}
	return checkFields("Network.IPAM", ipam, []string{fieldDriver}, nil)
}

func (p *Proxy) networkRead(c *call) {
	if err := p.usableNetwork(c.r.Context(), c.policy, c.args[0]); err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a network the app may use")
}

func (p *Proxy) networkDelete(c *call) {
	info, found, err := p.inspect(c.r.Context(), kindNetworks, c.args[0])
	if err == nil && (!found || info.Labels[OwnerLabel] != c.policy.AppID) {
		err = refusef("network %s is not one this app created", c.args[0])
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a network of the app's own")
}

func (p *Proxy) networkConnect(c *call) {
	ctx := c.r.Context()
	body, err := readBody(c.r)
	if err == nil {
		err = checkFields("connect", body, networkConnectFields, nil)
	}
	if err == nil {
		err = checkFields("connect.EndpointConfig", object(body["EndpointConfig"]), endpointFields, nil)
	}
	if err == nil {
		err = p.usableNetwork(ctx, c.policy, c.args[0])
	}
	if err == nil {
		err = p.childOrSelf(ctx, c.policy, text(body["Container"]))
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	p.forward(c, "a network the app may use")
}

// usableNetwork checks that children may join the network: the app's own, one
// the policy names, or one the app created.
func (p *Proxy) usableNetwork(ctx context.Context, policy *Policy, idOrName string) error {
	if policy.joinable(idOrName) {
		return nil
	}
	info, found, err := p.inspect(ctx, kindNetworks, idOrName)
	if err != nil {
		return err
	}
	if !found {
		return refusef("network %s does not exist", idOrName)
	}
	if policy.joinable(info.Name) || info.Labels[OwnerLabel] == policy.AppID {
		return nil
	}
	return refusef("network %s is not one this app may use", idOrName)
}

// childOrSelf also lets through the app's own task containers, which may
// connect themselves to the networks of their children.
func (p *Proxy) childOrSelf(ctx context.Context, policy *Policy, id string) error {
	labels, err := p.containerLabels(ctx, id)
	if err != nil {
		return err
	}
	if labels[OwnerLabel] == policy.AppID {
		return nil
	}
	if labels[OwnerLabel] == "" && policy.ServiceID != "" && labels[serviceIDLabel] == policy.ServiceID {
		return nil
	}
	return refusef("container %s is not one this app started", id)
}
