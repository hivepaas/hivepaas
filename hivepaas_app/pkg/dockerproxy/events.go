package dockerproxy

import (
	"encoding/json"
	"slices"
	"sort"
)

const (
	eventFilterContainer = "container"
	eventFilterType      = "type"
	eventTypeContainer   = "container"
)

// eventFilterKeys are the filters an event stream may carry. Each only narrows
// what a stream of the app's children says.
var eventFilterKeys = []string{eventFilterContainer, eventFilterType, "event"}

// events passes a stream of the events of the app's children, and nothing wider:
// the client names every container it wants, each must be a child, and events
// of containers are all it may ask for. The daemon applies the filter, so the
// stream holds what it names and nothing else. Appwrite's orchestrator watches
// each build's two containers this way, and falls back to asking every second
// when it cannot.
func (p *Proxy) events(c *call) {
	filters, err := eventFilters(c.r.URL.Query().Get("filters"))
	if err == nil {
		err = checkEventFilters(filters)
	}
	for _, id := range filters[eventFilterContainer] {
		if err != nil {
			break
		}
		err = p.child(c.r.Context(), c.policy, id)
	}
	if err != nil {
		p.refuse(c, err)
		return
	}
	p.forward(c, "events of the app's children")
}

// eventFilters reads the filters of an events request, in either of the two
// forms Docker takes them: {"type":["container"]}, and the older
// {"type":{"container":true}} the Go SDK still sends.
func eventFilters(raw string) (map[string][]string, error) {
	if raw == "" {
		return map[string][]string{}, nil
	}
	var lists map[string][]string
	if err := json.Unmarshal([]byte(raw), &lists); err == nil {
		return lists, nil
	}
	var sets map[string]map[string]bool
	if err := json.Unmarshal([]byte(raw), &sets); err != nil {
		return nil, refusef("events filters are not readable")
	}
	lists = make(map[string][]string, len(sets))
	for key, set := range sets {
		for value, on := range set {
			if on {
				lists[key] = append(lists[key], value)
			}
		}
		sort.Strings(lists[key])
	}
	return lists, nil
}

// checkEventFilters refuses a stream that could say anything of what is not the
// app's: one naming no container, or asking for more than container events.
func checkEventFilters(filters map[string][]string) error {
	for key := range filters {
		if !slices.Contains(eventFilterKeys, key) {
			return refusef("events filter %s is not allowed", key)
		}
	}
	if len(filters[eventFilterContainer]) == 0 {
		return refusef("events must be filtered to containers of the app")
	}
	if !slices.Equal(filters[eventFilterType], []string{eventTypeContainer}) {
		return refusef("events must be filtered to type container")
	}
	return nil
}
