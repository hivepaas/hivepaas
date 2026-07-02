package validation

func IsCommitHash(hash string) bool {
	if len(hash) != 40 && len(hash) != 64 { // SHA1: 40, SHA256: 64
		return false
	}
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) { //nolint
			return false
		}
	}
	return true
}
