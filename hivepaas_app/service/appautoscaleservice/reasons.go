package appautoscaleservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

// ReasonText is why what autoscale reads cannot be read, as a person reads it:
// for an error, or a tool's answer. The dashboard words them itself.
func ReasonText(reason string) string {
	switch reason {
	case string(loggingservice.HistoryReasonDisabled):
		return "stored logs are off; an administrator turns them on in System → Logging"
	case string(loggingservice.HistoryReasonAppsNotCollected):
		return "app logs are not collected; an administrator turns them on in System → Logging"
	case string(loggingservice.HistoryReasonNoQueryEndpoint):
		return "logs go to an external backend HivePaaS has no query endpoint for"
	case string(loggingservice.HistoryReasonDriverUnreadable):
		return "its log driver cannot be collected; switch it to json-file in its container settings"
	case string(loggingservice.HistoryReasonIdentityMissing):
		return "its containers do not carry its identity yet; save its container settings once"
	case ReasonNotExposed:
		return "it has no domain, so none of its requests go through Traefik"
	case string(traefikservice.AccessLogOff):
		return "Traefik's access log is off; an administrator turns it on in System → Traefik → Config Options"
	case string(traefikservice.AccessLogNotJSON), string(traefikservice.AccessLogUnlabelled):
		return "Traefik's access log is written in an older form; an administrator saves System → Traefik → " +
			"Config Options once, with Access Log on, and Traefik restarts briefly"
	case ReasonAgentUnlabelled:
		return "the HivePaaS agent does not mark its rows yet; it does from its next update, when HivePaaS is updated"
	case ReasonNoCPULimit:
		return "it has neither a CPU limit nor a CPU reservation to measure its CPU by; set one in its Resources"
	}
	return reason
}
