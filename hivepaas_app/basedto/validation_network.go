package basedto

import (
	"fmt"
	"net/netip"
	"strings"

	vld "github.com/tiendc/go-validator"
)

func ValidateDomain[T ~string](s *T, required bool, maxLen int, wildcardAllowed bool, field string) (
	result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		domain := string(*s)
		isWildcard := strings.HasPrefix(domain, "*.")
		if isWildcard {
			domain = strings.TrimPrefix(domain, "*.")
		}
		result = append(result,
			vld.StrLen(s, 1, maxLen).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
			),
			vld.StrIsDNSName(&domain).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_DOMAIN_INVALID"),
			),
		)
		if isWildcard && !wildcardAllowed {
			result = append(result, vld.Must(false).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_WILDCARD_UNALLOWED"),
			))
		}
	}
	return result
}

func ValidatePort[T int | uint | int32 | uint32 | int64 | uint64](v *T, required bool, min T,
	field string) []vld.Validator {
	return ValidateNumber(v, required, min, 65535, field) //nolint:mnd
}

// ValidateIPOrCIDRSlice checks every entry parses as an IP address or a CIDR block.
//
// These values reach Traefik as a command-line argument. An entry Traefik cannot
// parse is not a field-level annoyance there: the proxy refuses the argument and
// the ingress does not come up, so a typo here takes the whole install offline.
// Catching it at the edge is the difference between a form error and an outage.
func ValidateIPOrCIDRSlice(values []string, minLen int, field string) (result []vld.Validator) {
	result = append(result, ValidateSlice(values, true, minLen, nil, field)...)

	for i, value := range values {
		if isIPOrCIDR(value) {
			continue
		}
		result = append(result, vld.Must(false).OnError(
			vld.SetField(fmt.Sprintf("%s[%d]", field, i), nil),
			vld.SetCustomKey("ERR_VLD_IP_OR_CIDR_INVALID"),
		))
	}
	return result
}

func isIPOrCIDR(value string) bool {
	if _, err := netip.ParseAddr(value); err == nil {
		return true
	}
	_, err := netip.ParsePrefix(value)
	return err == nil
}
