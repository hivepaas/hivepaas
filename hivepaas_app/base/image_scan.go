package base

// VulnSeverity is how bad a vulnerability is, as Trivy grades it.
type VulnSeverity string

const (
	VulnSeverityCritical VulnSeverity = "CRITICAL"
	VulnSeverityHigh     VulnSeverity = "HIGH"
	VulnSeverityMedium   VulnSeverity = "MEDIUM"
	VulnSeverityLow      VulnSeverity = "LOW"
	VulnSeverityUnknown  VulnSeverity = "UNKNOWN"
)

var (
	// AllVulnSeverities run from the worst to the least.
	AllVulnSeverities = []VulnSeverity{VulnSeverityCritical, VulnSeverityHigh, VulnSeverityMedium,
		VulnSeverityLow, VulnSeverityUnknown}
)

// Rank is the severity's place among AllVulnSeverities: 0 for the worst. One
// Trivy does not name ranks with UNKNOWN.
func (s VulnSeverity) Rank() int {
	for i, severity := range AllVulnSeverities {
		if severity == s {
			return i
		}
	}
	return len(AllVulnSeverities) - 1
}

// ImageScanStatus is how an image's last scan ended.
type ImageScanStatus string

const (
	ImageScanStatusDone   ImageScanStatus = "done"
	ImageScanStatusFailed ImageScanStatus = "failed"
)

// ImageScanTrigger is what asked for a scan.
type ImageScanTrigger string

const (
	// ImageScanTriggerDeploy is a deployment that ended done.
	ImageScanTriggerDeploy ImageScanTrigger = "deploy"
	// ImageScanTriggerManual is somebody's Scan now.
	ImageScanTriggerManual ImageScanTrigger = "manual"
	// ImageScanTriggerDaily is the daily scan.
	ImageScanTriggerDaily ImageScanTrigger = "daily"
)
