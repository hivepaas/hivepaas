package dockerproxy

import (
	"fmt"
	"net/http"
	"strings"
)

// infoClusterFields are what /info says about the cluster and the node's place
// in it, rather than about the engine an app is talking to.
var infoClusterFields = []string{"Swarm", "Labels", "RegistryConfig"}

func (p *Proxy) pass(c *call) {
	p.forward(c, "read-only endpoint")
}

func (p *Proxy) info(c *call) {
	var answer map[string]any
	if !p.fetch(c, &answer) {
		return
	}
	for _, field := range infoClusterFields {
		delete(answer, field)
	}
	p.decide(c, true, "info, without the cluster")
	writeJSON(c.w, http.StatusOK, answer)
}

func (p *Proxy) pull(c *call) {
	query, err := formValues(c.r)
	if err != nil {
		p.refuse(c, err)
		return
	}
	if query.Get("fromSrc") != "" {
		p.refuse(c, refusef("importing an image is not allowed"))
		return
	}
	ref := query.Get("fromImage")
	if tag := query.Get("tag"); tag != "" {
		separator := ":"
		if strings.HasPrefix(tag, "sha256:") {
			separator = "@"
		}
		ref += separator + tag
	}
	if !matchImage(c.policy.Images, ref) {
		p.refuse(c, refusef("image %s is not allowed", ref))
		return
	}
	p.forward(c, "image "+ref)
}

// fetch reads what the client asked for from the daemon, for the proxy to trim
// before answering. It answers the client itself when that fails.
func (p *Proxy) fetch(c *call, out any) bool {
	found, err := p.daemon.get(c.r.Context(), c.r.URL.RequestURI(), out)
	if err == nil && !found {
		err = fmt.Errorf("%w: GET %s answered 404", errDaemon, c.r.URL.Path)
	}
	if err != nil {
		p.refuse(c, err)
		return false
	}
	return true
}
