package dockerhelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/services/docker"
)

var (
	restrictedSystemLabels = map[string]struct{}{
		docker.StackLabelNamespace: {},
	}

	restrictedLabelPrefixes = []string{"hivepaas.", "traefik."}
)

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
	if _, exists := restrictedSystemLabels[labelKey]; exists {
		return true
	}
	for _, prefix := range restrictedLabelPrefixes {
		if strings.HasPrefix(labelKey, prefix) {
			return true
		}
	}
	return false
}
