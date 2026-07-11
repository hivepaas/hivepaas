package dockerhelper

import "strings"

func WrapNodeID(id string) string {
	return "dkr:node:" + id
}

func WrapNetworkID(id string) string {
	return "dkr:net:" + id
}

func WrapVolumeID(id string) string {
	return "dkr:vol:" + id
}

func ParseID(wrapID string) string {
	wrapID = strings.TrimPrefix(wrapID, "dkr:")
	before, after, found := strings.Cut(wrapID, ":")
	if !found {
		return wrapID
	}
	switch before {
	case "node", "net", "vol":
		return after
	default:
		return wrapID
	}
}
