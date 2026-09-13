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
