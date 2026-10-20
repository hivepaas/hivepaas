package entity

import (
	"encoding/json"
)

// A step of a setting's Migrate brings a row from the shape of the version it
// names to the next one's. Until data migration was fixed it did not run, and
// the settings screens saved rows in the current shape while keeping the
// version they were made at. So a step that sets a value must tell the old
// shape by the data, not by the version alone: what a row already has, its
// person may have chosen.

// storedHas says whether the row's data, as stored, has the field at path: an
// object's keys, one level each.
func storedHas(data string, path ...string) bool {
	raw := json.RawMessage(data)
	for _, key := range path {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			return false
		}
		next, found := object[key]
		if !found {
			return false
		}
		raw = next
	}
	return true
}
