package obi

// Capacity is how many requests and connections OBI tracks at once on a node:
// the size of its eBPF maps, which it allocates whole when it starts. More
// costs memory, idle or not; too little loses what does not fit, silently.
type Capacity string

const (
	// CapacityAuto is the capacity HivePaaS recommends for the node's memory.
	CapacityAuto Capacity = "auto"
	// CapacitySmall is a quarter of OBI's maps (global_scale_factor -2):
	// about 100 MiB, enough for a node that is not serving thousands of
	// requests or connections at once.
	CapacitySmall Capacity = "small"
	// CapacityMedium is half (-1): about 140 MiB, twice as many.
	CapacityMedium Capacity = "medium"
	// CapacityLarge is OBI's own default (0): about 215 MiB, four times as many.
	CapacityLarge Capacity = "large"
)

// level is a capacity's maps, and what it was measured at (step 0, OBI
// v0.14.0, a 1-vCPU node): its memory, and the in-flight entries of each of
// its maps. A smaller one (-3) lost a third of the requests.
type level struct {
	scaleFactor int
	memoryMiB   int
	tracked     int
}

var levels = map[Capacity]level{
	CapacitySmall:  {scaleFactor: -2, memoryMiB: 100, tracked: 7500},  //nolint:mnd // measured
	CapacityMedium: {scaleFactor: -1, memoryMiB: 140, tracked: 15000}, //nolint:mnd // measured
	CapacityLarge:  {scaleFactor: 0, memoryMiB: 215, tracked: 30000},  //nolint:mnd // measured
}

// Capacities are the ones a node can be given, smallest first.
var Capacities = []Capacity{CapacitySmall, CapacityMedium, CapacityLarge}

const (
	// mediumFromMB and largeFromMB are the node memory from which a bigger
	// capacity is recommended: a node with more memory spares it more easily,
	// and tends to serve more at once.
	mediumFromMB = 8 << 10
	largeFromMB  = 32 << 10
)

// ParseCapacity reads a node's setting: "" and "auto" are CapacityAuto. false
// for one that is no capacity.
func ParseCapacity(s string) (Capacity, bool) {
	switch c := Capacity(s); c {
	case "", CapacityAuto:
		return CapacityAuto, true
	case CapacitySmall, CapacityMedium, CapacityLarge:
		return c, true
	}
	return "", false
}

// Recommended is the capacity HivePaaS recommends for a node of memTotalMB:
// small under 8 GB, medium under 32 GB, large from 32 GB. Small when the
// memory is unknown.
func Recommended(memTotalMB int) Capacity {
	switch {
	case memTotalMB >= largeFromMB:
		return CapacityLarge
	case memTotalMB >= mediumFromMB:
		return CapacityMedium
	}
	return CapacitySmall
}

// Effective is the capacity a node runs with: the one chosen, or for auto the
// one recommended for its memory.
func (c Capacity) Effective(memTotalMB int) Capacity {
	if _, ok := levels[c]; ok {
		return c
	}
	return Recommended(memTotalMB)
}

// ScaleFactor is OBI's maps_config.global_scale_factor for a capacity.
func (c Capacity) ScaleFactor() int { return levels[c.Effective(0)].scaleFactor }

// MemoryMiB is about what OBI takes at a capacity, its maps included.
func (c Capacity) MemoryMiB() int { return levels[c.Effective(0)].memoryMiB }

// Tracked is about how many requests and connections OBI tracks at once at a
// capacity: the entries of each of its in-flight maps.
func (c Capacity) Tracked() int { return levels[c.Effective(0)].tracked }
