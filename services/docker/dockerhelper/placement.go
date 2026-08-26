package dockerhelper

import "strings"

// ParsePlacementConstraint parses a Docker Swarm placement constraint string
// into (key, op, value). Supported operators: "==" and "!=".
// If the constraint is invalid (no operator found or key is empty), it returns ("", "", "").
func ParsePlacementConstraint(s string) (k, op, v string) {
	op = "=="
	kk, vv, found := strings.Cut(s, op)
	if !found {
		op = "!="
		kk, vv, found = strings.Cut(s, op)
	}
	if !found {
		return "", "", ""
	}
	k, v = strings.TrimSpace(kk), strings.TrimSpace(vv)
	if k == "" {
		return "", "", ""
	}
	return k, op, v
}
