package dockerproxy

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// minCreateMajor and minCreateMinor are the API version a rewritten create
// needs: a mount's VolumeOptions.Subpath came in 1.45.
const (
	minCreateMajor = 1
	minCreateMinor = 45
)

const (
	networkModeNone      = "none"
	networkModeContainer = "container:"
)

var (
	// configFields are the fields of a container's Config a child may set.
	configFields = []string{"Hostname", "Domainname", "User", "AttachStdin", "AttachStdout", "AttachStderr",
		"ExposedPorts", "Tty", "OpenStdin", "StdinOnce", "Env", "Cmd", "Healthcheck", "ArgsEscaped", "Image",
		"Volumes", "WorkingDir", "Entrypoint", "NetworkDisabled", fieldLabels, "StopSignal", "StopTimeout",
		"Shell", "HostConfig", "NetworkingConfig"}
	// hostFields are the fields of a HostConfig a child may set. Everything that
	// reaches past the container - Privileged, CapAdd, Devices, the host's
	// namespaces, SecurityOpt, Runtime, Sysctls, CgroupParent, VolumesFrom,
	// Links, PortBindings - is absent, and so refused.
	hostFields = []string{"NetworkMode", "RestartPolicy", "AutoRemove", "Memory", "MemorySwap",
		"MemoryReservation", "NanoCpus", "CpuQuota", "CpuPeriod", "CpuShares", "PidsLimit", "ShmSize", "Dns",
		"DnsOptions", "DnsSearch", "ExtraHosts", "LogConfig", "Init", "ReadonlyRootfs", "Tmpfs", "CapDrop",
		"Ulimits", "GroupAdd", "ConsoleSize", "Isolation"}
	// hostNullOnly are fields whose empty list unmasks /proc.
	hostNullOnly     = []string{"MaskedPaths", "ReadonlyPaths"}
	networkingFields = []string{"EndpointsConfig"}
	// logDrivers write where docker logs reads, and nowhere else.
	logDrivers = []string{"", "json-file", local}
	// defaultNetworkModes are what a client asks for when it names no network of
	// its own, and they all become the app's network. host is among them because
	// the apps that ask for it - Autobase's automation - want the internet, which
	// the app's network reaches too.
	defaultNetworkModes = []string{"", "default", "bridge", "host"}
)

func (p *Proxy) create(c *call) {
	body, err := readBody(c.r)
	if err == nil {
		err = p.checkCreate(c.r.Context(), c.policy, body)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	raiseVersion(c.r)
	p.forward(c, "a child of the app")
}

// checkCreate judges a container create and rewrites it into what the app may
// have. It changes body in place.
func (p *Proxy) checkCreate(ctx context.Context, policy *Policy, body map[string]any) error {
	if err := checkFields("Config", body, configFields, nil); err != nil {
		return err
	}
	if image := text(body["Image"]); !matchImage(policy.Images, image) {
		return refusef("image %s is not allowed", image)
	}
	host := object(body["HostConfig"])
	if host == nil {
		host = map[string]any{}
		body["HostConfig"] = host
	}
	if err := checkFields("HostConfig", host, hostFields, hostNullOnly); err != nil {
		return err
	}
	if driver := text(object(host["LogConfig"])["Type"]); !slices.Contains(logDrivers, driver) {
		return refusef("log driver %s is not allowed", driver)
	}
	if err := p.rewriteNetwork(ctx, policy, body, host); err != nil {
		return err
	}
	if err := applyLimits(host, policy.Limits); err != nil {
		return err
	}
	body[fieldLabels] = ownLabels(object(body[fieldLabels]), policy.AppID)
	return p.checkCount(ctx, policy)
}

// rewriteNetwork puts a child on the app's network unless it names another the
// app may use.
func (p *Proxy) rewriteNetwork(ctx context.Context, policy *Policy, body, host map[string]any) error {
	mode := text(host["NetworkMode"])
	switch {
	case slices.Contains(defaultNetworkModes, mode):
		host["NetworkMode"] = policy.Network
	case mode == networkModeNone:
	case strings.HasPrefix(mode, networkModeContainer):
		return refusef("sharing another container's network is not allowed")
	default:
		if err := p.usableNetwork(ctx, policy, mode); err != nil {
			return err
		}
	}
	networking := object(body["NetworkingConfig"])
	if err := checkFields("NetworkingConfig", networking, networkingFields, nil); err != nil {
		return err
	}
	endpoints := object(networking["EndpointsConfig"])
	for _, name := range slices.Sorted(maps.Keys(endpoints)) {
		if err := checkFields("EndpointsConfig."+name, object(endpoints[name]), endpointFields, nil); err != nil {
			return err
		}
		if slices.Contains(defaultNetworkModes, name) {
			endpoints[policy.Network] = endpoints[name]
			delete(endpoints, name)
			continue
		}
		if err := p.usableNetwork(ctx, policy, name); err != nil {
			return err
		}
	}
	return nil
}

func (p *Proxy) checkCount(ctx context.Context, policy *Policy) error {
	var children []struct {
		ID string `json:"Id"`
	}
	query := "/containers/json?all=1&filters=" + labelFilter(OwnerLabel, policy.AppID)
	if _, err := p.daemon.get(ctx, query, &children); err != nil {
		return err
	}
	if len(children) >= policy.Limits.Containers {
		return refusef("this app already has %d containers, its limit", len(children))
	}
	return nil
}

// raiseVersion sends a create at the version it needs when the client asked for
// an older one. What the client sent means the same at the higher version, and a
// request with no version already gets the daemon's own.
func raiseVersion(r *http.Request) {
	current := versionPrefix.FindString(r.URL.Path)
	if current == "" {
		return
	}
	major, minor := splitVersion(strings.TrimPrefix(current, "/v"))
	if major > minCreateMajor || (major == minCreateMajor && minor >= minCreateMinor) {
		return
	}
	r.URL.Path = fmt.Sprintf("/v%d.%d", minCreateMajor, minCreateMinor) + strings.TrimPrefix(r.URL.Path, current)
	r.URL.RawPath = ""
}

func splitVersion(v string) (int, int) {
	major, minor, _ := strings.Cut(v, ".")
	majorNumber, _ := strconv.Atoi(major)
	minorNumber, _ := strconv.Atoi(minor)
	return majorNumber, minorNumber
}
