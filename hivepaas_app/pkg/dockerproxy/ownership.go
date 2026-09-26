package dockerproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// fieldLabels is where Docker keeps an object's labels.
const fieldLabels = "Labels"

// execFields are what an exec in a child may ask for. Privileged is not among
// them.
var execFields = []string{"User", "Cmd", "Env", "WorkingDir", "Tty", "AttachStdin", "AttachStdout",
	"AttachStderr", "DetachKeys", "ConsoleSize"}

type containerInspect struct {
	Config struct {
		Labels map[string]string
	}
}

// containerLabels returns a container's labels, refusing one that does not exist.
func (p *Proxy) containerLabels(ctx context.Context, id string) (map[string]string, error) {
	var info containerInspect
	found, err := p.daemon.get(ctx, "/containers/"+url.PathEscape(id)+"/json", &info)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, refusef("container %s does not exist", id)
	}
	return info.Config.Labels, nil
}

// child checks that a container is one the app started.
func (p *Proxy) child(ctx context.Context, policy *Policy, id string) error {
	labels, err := p.containerLabels(ctx, id)
	if err != nil {
		return err
	}
	if labels[OwnerLabel] != policy.AppID {
		return refusef("container %s is not one this app started", id)
	}
	return nil
}

// ownedBy reports whether labels, as a list answer carries them, mark the app's.
func ownedBy(labels any, appID string) bool {
	return text(object(labels)[OwnerLabel]) == appID
}

// ownTask reports whether labels, as a list answer carries them, mark one of the
// app's own task containers. A container carrying the owner label is a child,
// whatever else it says, as taskMounts has it.
func ownTask(labels any, policy *Policy) bool {
	fields := object(labels)
	return policy.ServiceID != "" && text(fields[serviceIDLabel]) == policy.ServiceID && text(fields[OwnerLabel]) == ""
}

// keep returns the items ok allows, and an empty list rather than none.
func keep(items []map[string]any, ok func(map[string]any) bool) []map[string]any {
	kept := []map[string]any{}
	for _, item := range items {
		if ok(item) {
			kept = append(kept, item)
		}
	}
	return kept
}

func (p *Proxy) listContainers(c *call) {
	var items []map[string]any
	if !p.fetch(c, &items) {
		return
	}
	// The app's own tasks are listed beside its children, so that it can find
	// itself: Appwrite's executor looks itself up by its host name, to connect
	// itself to the network of its runtimes. Listing is all it is given of them;
	// everything else it names must still be a child.
	kept := keep(items, func(item map[string]any) bool {
		return ownedBy(item[fieldLabels], c.policy.AppID) || ownTask(item[fieldLabels], c.policy)
	})
	p.decide(c, true, fmt.Sprintf("%d of %d containers are the app's or its own", len(kept), len(items)))
	writeJSON(c.w, http.StatusOK, kept)
}

func (p *Proxy) onChild(c *call) {
	if err := p.child(c.r.Context(), c.policy, c.args[0]); err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "a child of the app")
}

func (p *Proxy) onExec(c *call) {
	ctx := c.r.Context()
	var exec struct {
		ContainerID string
	}
	found, err := p.daemon.get(ctx, "/exec/"+url.PathEscape(c.args[0])+"/json", &exec)
	if err == nil && !found {
		err = refusef("exec %s does not exist", c.args[0])
	}
	if err == nil {
		err = p.child(ctx, c.policy, exec.ContainerID)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "an exec in a child of the app")
}

func (p *Proxy) execCreate(c *call) {
	err := p.child(c.r.Context(), c.policy, c.args[0])
	var body map[string]any
	if err == nil {
		body, err = readBody(c.r)
	}
	if err == nil {
		err = checkFields("exec", body, execFields, nil)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	setBody(c.r, body)
	p.forward(c, "an exec in a child of the app")
}
