package base

// LoggingBackendType names the kind of log store, independently of who runs it.
type LoggingBackendType string

const (
	LoggingBackendTypeVictoriaLogs LoggingBackendType = "victoria-logs"
)

var (
	AllLoggingBackendTypes = []LoggingBackendType{LoggingBackendTypeVictoriaLogs}
)

// LoggingCollectorType names the kind of collector, independently of who runs it.
type LoggingCollectorType string

const (
	LoggingCollectorTypeVlagent LoggingCollectorType = "vlagent"
)

var (
	AllLoggingCollectorTypes = []LoggingCollectorType{LoggingCollectorTypeVlagent}
)

// LabelLogComponent is the container label naming which of HivePaaS's own
// services a container is. As an app's id is, it is copied into every line the
// container writes by the json-file driver's `labels` option - by the daemon,
// not by the container - which is what lets a query trust that a line is, say,
// the proxy's and not one an app printed to look like it.
const LabelLogComponent = "hivepaas.component"

// LogComponentTraefik is the proxy's LabelLogComponent: its access log is
// where an app's HTTP numbers are counted from.
const LogComponentTraefik = "traefik"

// LogComponentAgent is the agent's LabelLogComponent: the containers' CPU and
// memory are counted from the rows it writes.
const LogComponentAgent = "agent"

// LogComponentOBI is the LabelLogComponent of OBI, which the agent runs on a
// node to see apps' routes and calls: its own log lines are told apart by it.
const LogComponentOBI = "obi"
