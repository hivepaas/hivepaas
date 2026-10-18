package base

// CPUArch is a node's processor architecture, spelled as Go and image platforms
// spell it. A downloaded binary is built for one, so anything HivePaaS puts on a
// node by itself has to be chosen by it.
type CPUArch string

const (
	CPUArchAMD64 CPUArch = "amd64"
	CPUArchARM64 CPUArch = "arm64"
)

var AllCPUArches = []CPUArch{CPUArchAMD64, CPUArchARM64}

// CPUArchFromDocker reads the architecture a Swarm node reports. Docker gives
// it uname's way - x86_64, aarch64 - in Description.Platform.Architecture, and
// some engines give Go's; both are accepted. An architecture HivePaaS has no
// build for is not.
func CPUArchFromDocker(reported string) (CPUArch, bool) {
	switch reported {
	case "x86_64", string(CPUArchAMD64):
		return CPUArchAMD64, true
	case "aarch64", string(CPUArchARM64):
		return CPUArchARM64, true
	default:
		return "", false
	}
}
