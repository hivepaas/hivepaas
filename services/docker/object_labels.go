package docker

import (
	"strings"
)

const (
	restrictedLabelsPrefix = "localpaas."
)

var (
	restrictedSystemLabels = map[string]struct{}{
		StackLabelNamespace: {},
	}
)

func FilterOutSystemLabels(labels map[string]string) map[string]string {
	resp := make(map[string]string, len(labels))
	for k, v := range labels {
		if _, exists := restrictedSystemLabels[k]; exists {
			continue
		}
		if strings.HasPrefix(k, restrictedLabelsPrefix) {
			continue
		}
		resp[k] = v
	}
	return resp
}

func ValidateUserLabels(labels map[string]string, stopAtFirstViolation bool) (unallowedLabels []string) {
	for k := range labels {
		if _, exists := restrictedSystemLabels[k]; exists {
			if stopAtFirstViolation {
				return []string{k}
			}
			unallowedLabels = append(unallowedLabels, k)
			continue
		}
		if strings.HasPrefix(k, restrictedLabelsPrefix) {
			if stopAtFirstViolation {
				return []string{k}
			}
			unallowedLabels = append(unallowedLabels, k)
			continue
		}
	}
	return unallowedLabels
}

func ApplyUserLabels(currLabels, userLabels map[string]string) map[string]string {
	appliedLabels := make(map[string]string, len(userLabels))
	for k, v := range currLabels {
		if _, exists := restrictedSystemLabels[k]; exists {
			appliedLabels[k] = v
			continue
		}
		if strings.HasPrefix(k, restrictedLabelsPrefix) {
			appliedLabels[k] = v
			continue
		}
	}
	for k, v := range userLabels {
		if _, exists := restrictedSystemLabels[k]; exists {
			continue
		}
		if strings.HasPrefix(k, restrictedLabelsPrefix) {
			continue
		}
		appliedLabels[k] = v
	}
	return appliedLabels
}
