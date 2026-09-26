package dockerproxy

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
)

type fakeContainer struct {
	labels  map[string]string
	mounts  []map[string]any
	running bool
}

type fakeNetwork struct {
	name   string
	labels map[string]string
}

type recordedRequest struct {
	method string
	path   string
	query  string
	body   []byte
}

// fakeDaemon answers the lookups the proxy makes from a world a test sets up,
// and records every request that reaches it.
type fakeDaemon struct {
	mu         sync.Mutex
	containers map[string]*fakeContainer
	execs      map[string]string
	volumes    map[string]map[string]string
	// volumeInfo is what inspecting a volume says besides its name and labels.
	volumeInfo map[string]map[string]any
	networks   map[string]*fakeNetwork
	requests   []recordedRequest
}

func (f *fakeDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	path := versionPrefix.ReplaceAllString(r.URL.Path, "")
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{r.Method, r.URL.Path, r.URL.RawQuery, body})
	f.mu.Unlock()
	if r.Method == http.MethodPost && strings.HasSuffix(path, "/attach") {
		f.attach(w)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case r.Method == http.MethodGet && path == "/containers/json":
		f.listContainers(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/containers/") && strings.HasSuffix(path, "/json"):
		f.inspectContainer(w, strings.TrimSuffix(strings.TrimPrefix(path, "/containers/"), "/json"))
	case r.Method == http.MethodPost && path == "/containers/create":
		writeJSON(w, http.StatusCreated, map[string]any{"Id": "created", "Warnings": []string{}})
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/exec/"):
		f.inspectExec(w, strings.TrimSuffix(strings.TrimPrefix(path, "/exec/"), "/json"))
	case r.Method == http.MethodGet && path == "/volumes":
		f.listVolumes(w)
	case r.Method == http.MethodPost && path == "/volumes/create":
		f.createVolume(w, body)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/volumes/"):
		f.inspectVolume(w, strings.TrimPrefix(path, "/volumes/"))
	case r.Method == http.MethodGet && path == "/networks":
		f.listNetworks(w)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/networks/"):
		f.inspectNetwork(w, strings.TrimPrefix(path, "/networks/"))
	case r.Method == http.MethodGet && path == "/info":
		writeJSON(w, http.StatusOK, map[string]any{
			"ServerVersion": "29.8.0", "OSType": "linux",
			"Swarm": map[string]any{"NodeID": "node1"}, "Labels": []string{"zone=a"},
			"RegistryConfig": map[string]any{"Mirrors": []string{}},
		})
	default:
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

func (f *fakeDaemon) listContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1"
	var filters map[string][]string
	_ = json.Unmarshal([]byte(r.URL.Query().Get("filters")), &filters)
	out := []map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(f.containers)) {
		c := f.containers[id]
		if (all || c.running) && hasLabels(c.labels, filters["label"]) {
			out = append(out, map[string]any{"Id": id, "Labels": c.labels})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func hasLabels(labels map[string]string, wanted []string) bool {
	for _, pair := range wanted {
		key, value, _ := strings.Cut(pair, "=")
		if labels[key] != value {
			return false
		}
	}
	return true
}

func (f *fakeDaemon) inspectContainer(w http.ResponseWriter, id string) {
	c, found := f.containers[id]
	if !found {
		writeError(w, http.StatusNotFound, "No such container: "+id)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"Id": id, "Config": map[string]any{"Labels": c.labels}, "HostConfig": map[string]any{"Mounts": c.mounts},
	})
}

func (f *fakeDaemon) inspectExec(w http.ResponseWriter, id string) {
	container, found := f.execs[id]
	if !found {
		writeError(w, http.StatusNotFound, "No such exec instance: "+id)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ID": id, "ContainerID": container})
}

func (f *fakeDaemon) listVolumes(w http.ResponseWriter) {
	volumes := []map[string]any{}
	for _, name := range slices.Sorted(maps.Keys(f.volumes)) {
		volumes = append(volumes, map[string]any{"Name": name, "Labels": f.volumes[name]})
	}
	writeJSON(w, http.StatusOK, map[string]any{"Volumes": volumes, "Warnings": nil})
}

func (f *fakeDaemon) createVolume(w http.ResponseWriter, body []byte) {
	var req struct {
		Name   string
		Labels map[string]string
	}
	_ = json.Unmarshal(body, &req)
	f.volumes[req.Name] = req.Labels
	writeJSON(w, http.StatusCreated, map[string]any{"Name": req.Name, "Labels": req.Labels})
}

func (f *fakeDaemon) inspectVolume(w http.ResponseWriter, name string) {
	labels, found := f.volumes[name]
	if !found {
		writeError(w, http.StatusNotFound, "no such volume")
		return
	}
	info := map[string]any{"Name": name, "Labels": labels}
	maps.Copy(info, f.volumeInfo[name])
	writeJSON(w, http.StatusOK, info)
}

func (f *fakeDaemon) listNetworks(w http.ResponseWriter) {
	out := []map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(f.networks)) {
		n := f.networks[id]
		out = append(out, map[string]any{"Id": id, "Name": n.name, "Labels": n.labels})
	}
	writeJSON(w, http.StatusOK, out)
}

func (f *fakeDaemon) inspectNetwork(w http.ResponseWriter, idOrName string) {
	for id, n := range f.networks {
		if id == idOrName || n.name == idOrName {
			writeJSON(w, http.StatusOK, map[string]any{"Id": id, "Name": n.name, "Labels": n.labels})
			return
		}
	}
	writeError(w, http.StatusNotFound, "network "+idOrName+" not found")
}

// attach answers the way the daemon does: it takes the connection over and
// streams.
func (f *fakeDaemon) attach(w http.ResponseWriter) {
	conn, buf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	_, _ = buf.WriteString("HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\n" +
		"Connection: Upgrade\r\nUpgrade: tcp\r\n\r\nstream-ok\n")
	_ = buf.Flush()
}
