package dockerproxy

const (
	defaultPidsLimit = 1024
	maxPidsLimit     = 4096
	// A CPU period is what a quota is a share of, and what docker accepts for it.
	defaultCPUPeriod = 100_000
	minCPUPeriod     = 1_000
	maxCPUPeriod     = 1_000_000
	nanoPerCPU       = 1_000_000_000
)

// applyLimits gives a child the policy's limits when it asks for none, and
// refuses one that asks for more.
func applyLimits(host map[string]any, limits Limits) error {
	memory, err := number(host["Memory"])
	if err != nil {
		return err
	}
	switch {
	case memory <= 0:
		host["Memory"] = limits.Memory
	case memory > limits.Memory:
		return refusef("memory %d is more than the %d this app's children may have", memory, limits.Memory)
	}
	swap, err := number(host["MemorySwap"])
	if err != nil {
		return err
	}
	if swap < 0 {
		return refusef("unlimited swap is not allowed")
	}
	if err = applyCPU(host, limits.NanoCPUs); err != nil {
		return err
	}
	return applyPids(host)
}

// applyCPU caps processor time, given either way docker takes it: NanoCpus, or
// a quota of a period.
func applyCPU(host map[string]any, limit int64) error {
	quota, err := number(host["CpuQuota"])
	if err != nil {
		return err
	}
	if quota <= 0 {
		nano, err := number(host["NanoCpus"])
		if err != nil {
			return err
		}
		switch {
		case nano <= 0:
			host["NanoCpus"] = limit
		case nano > limit:
			return refusef("cpus %.2f is more than the %.2f this app's children may have",
				float64(nano)/nanoPerCPU, float64(limit)/nanoPerCPU)
		}
		return nil
	}
	period, err := number(host["CpuPeriod"])
	if err != nil {
		return err
	}
	if period == 0 {
		period = defaultCPUPeriod
	}
	if period < minCPUPeriod || period > maxCPUPeriod {
		return refusef("cpu period %d is outside %d-%d", period, minCPUPeriod, maxCPUPeriod)
	}
	// Compared as quotas: the limit times a period stays far inside an int64,
	// where the quota times a billion need not.
	if allowed := limit * period / nanoPerCPU; quota > allowed {
		return refusef("cpu quota %d of %d is more than the %.2f cpus this app's children may have",
			quota, period, float64(limit)/nanoPerCPU)
	}
	return nil
}

func applyPids(host map[string]any) error {
	pids, err := number(host["PidsLimit"])
	if err != nil {
		return err
	}
	switch {
	case pids <= 0:
		host["PidsLimit"] = int64(defaultPidsLimit)
	case pids > maxPidsLimit:
		return refusef("pids limit %d is more than %d", pids, maxPidsLimit)
	}
	return nil
}
