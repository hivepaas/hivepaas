package netutil

import "net/netip"

// IPAllowed reports whether an address falls inside any of the allowed entries,
// each of which is a single address or a CIDR block.
func IPAllowed(allowedIPs []string, clientIP string) bool {
	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return false
	}

	for _, allowed := range allowedIPs {
		if single, err := netip.ParseAddr(allowed); err == nil {
			if single == addr {
				return true
			}
			continue
		}
		if prefix, err := netip.ParsePrefix(allowed); err == nil && prefix.Contains(addr) {
			return true
		}
	}
	return false
}
