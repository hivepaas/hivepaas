package base

const (
	HivepaasGlobalBuilder = "hivepaas_builder"
)

type DockerfileSource string

const (
	DockerfileSourceManual DockerfileSource = "manual"
	DockerfileSourceAuto   DockerfileSource = "auto"
)

var (
	AllDockerfileSources = []DockerfileSource{DockerfileSourceManual, DockerfileSourceAuto}
)
