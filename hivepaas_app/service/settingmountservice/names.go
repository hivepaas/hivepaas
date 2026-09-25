package settingmountservice

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const (
	// LabelAppID is the app a mounted object belongs to, as elsewhere
	// (appservice.LabelLogAppID).
	LabelAppID = "hivepaas.app.id"
	// LabelEntry marks an object as a mounted setting's, with the entry's key.
	LabelEntry = "hivepaas.settingMount.entry"
	// LabelPart is the part the object holds.
	LabelPart = "hivepaas.settingMount.part"

	maxNameLen     = 64
	hashLen        = 8
	entryKeyMaxLen = 20
)

var (
	entryKeyPattern  = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,18}[a-z0-9])?$`)
	notEntryKeyChars = regexp.MustCompile(`[^a-z0-9]+`)
)

// ValidEntryKey reports whether an entry may be called key.
func ValidEntryKey(key string) bool {
	return entryKeyPattern.MatchString(key)
}

// EntryKeyFor is the entry key a setting's name makes, for the entries
// HivePaaS makes itself - a template's swarmRef, a system app's secret:
// lowercased, anything else a hyphen, at most 20 characters.
func EntryKeyFor(name string) string {
	key := strings.Trim(notEntryKeyChars.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(key) > entryKeyMaxLen {
		key = strings.Trim(key[:entryKeyMaxLen], "-")
	}
	return key
}

// ObjectName is the Docker secret or config a file is held in, named as secrets
// and config files are: GlobalKey + "_" + a name, lowercased. Docker caps names
// at 64 characters and a GlobalKey can be longer; such a name keeps as much of
// the GlobalKey as fits, and a hash of the whole of it.
func ObjectName(globalKey, entry, part, rotation string) string {
	suffix := strings.ToLower("_mount_" + entry + "_" + part + "_" + rotation[:min(hashLen, len(rotation))])
	prefix := strings.ToLower(globalKey)
	if len(prefix)+len(suffix) > maxNameLen {
		sum := sha256.Sum256([]byte(prefix))
		keep := max(0, maxNameLen-len(suffix)-hashLen)
		prefix = prefix[:min(keep, len(prefix))] + hex.EncodeToString(sum[:])[:hashLen]
	}
	return prefix + suffix
}

// Labels are those of a mounted object.
func Labels(appID, entry, part string) map[string]string {
	return map[string]string{LabelAppID: appID, LabelEntry: entry, LabelPart: part}
}
