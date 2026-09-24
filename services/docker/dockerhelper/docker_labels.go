package dockerhelper

import (
	"strings"
)

// restrictedLabelPrefixes are the labels HivePaaS and Docker manage - Docker's
// stack labels among them - which a person neither sees by default nor sets.
var restrictedLabelPrefixes = []string{"hivepaas.", "traefik.", "com.docker.stack."}

func FilterOutRestrictedLabels(labels map[string]string) map[string]string {
	resp := make(map[string]string, len(labels))
	for k, v := range labels {
		if isLabelRestricted(k) {
			continue
		}
		resp[k] = v
	}
	return resp
}

func ValidateUserLabels(labels map[string]string, stopAtFirstViolation bool) (unallowedLabels []string) {
	for k := range labels {
		if isLabelRestricted(k) {
			if stopAtFirstViolation {
				return []string{k}
			}
			unallowedLabels = append(unallowedLabels, k)
		}
	}
	return unallowedLabels
}

func ApplyUserLabels(currLabels, userLabels map[string]string) map[string]string {
	allowedLabels := make(map[string]string, len(userLabels))
	for k, v := range currLabels {
		if isLabelRestricted(k) {
			allowedLabels[k] = v
		}
	}
	for k, v := range userLabels {
		if isLabelRestricted(k) {
			continue
		}
		allowedLabels[k] = v
	}
	return allowedLabels
}

func isLabelRestricted(labelKey string) bool {
	labelKey = strings.ToLower(labelKey)
	for _, prefix := range restrictedLabelPrefixes {
		if strings.HasPrefix(labelKey, prefix) {
			return true
		}
	}
	return false
}
