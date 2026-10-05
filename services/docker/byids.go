package docker

import "strings"

// The ...ListByIDs functions list the objects of some ids in one call, and only
// those. Docker reads an id filter as a prefix, ORing the values of one filter,
// so it may answer more than was asked: what it answers is narrowed to the ids
// given. A blank id is dropped, since as a prefix it matches every object, and
// with none left Docker is not asked.

// wantedIDs is the ids asked for, each once, without the blank ones.
func wantedIDs(ids []string) map[string]struct{} {
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) != "" {
			wanted[id] = struct{}{}
		}
	}
	return wanted
}

// keepWanted keeps, in place, the items whose id is wanted.
func keepWanted[T any](items []T, wanted map[string]struct{}, idOf func(*T) string) []T {
	kept := items[:0]
	for i := range items {
		if _, ok := wanted[idOf(&items[i])]; ok {
			kept = append(kept, items[i])
		}
	}
	return kept
}
