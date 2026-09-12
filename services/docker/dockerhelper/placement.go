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

// NodeLabelConstraint turns a `key=value` node label selector into a placement
// constraint with the given operator ("==" or "!=").
//
// A bare key means `key=true`: that is how a label used as a flag is written,
// and how build-node exclusions have always been read.
//
// It reports false for what cannot be one, rather than emitting a constraint
// nobody meant: an empty key, a pasted operator (`k==v`, `k!=v`), or a comma -
// HivePaaS records the constraints it added in one comma-joined service label,
// so a comma inside a selector would come back as two selectors.
func NodeLabelConstraint(label, op string) (string, bool) {
	key, value, ok := ParseNodeLabelSelector(label)
	if !ok {
		return "", false
	}
	return "node.labels." + key + op + value, true
}

// ParseNodeLabelSelector reads a `key=value` node label selector, resolving a
// bare key to `key=true`. It is what NodeLabelConstraint is built from, and
// what deciding whether a node matches one reads.
func ParseNodeLabelSelector(label string) (key, value string, ok bool) {
	if strings.Contains(label, ",") {
		return "", "", false
	}
	k, v, found := strings.Cut(label, "=")
	k = strings.TrimSpace(k)
	v = strings.TrimSpace(v)
	if k == "" || strings.HasSuffix(k, "!") || strings.HasPrefix(v, "=") {
		return "", "", false
	}
	if !found || v == "" {
		v = "true"
	}
	return k, v, true
}
