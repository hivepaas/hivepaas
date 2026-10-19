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

// ImageScanSource is where a scan read its image.
type ImageScanSource string

const (
	// ImageScanSourceNode is the Docker of the node the scan ran on, which runs
	// the image.
	ImageScanSourceNode ImageScanSource = "node"
	// ImageScanSourceRegistry is the image's registry, by digest: the node the
	// scans run on does not have it.
	ImageScanSourceRegistry ImageScanSource = "registry"
	// ImageScanSourceArtifact is a file of an app's artifacts in place of an
	// image, read from the artifacts cache of the node the scan ran on.
	ImageScanSourceArtifact ImageScanSource = "artifact"
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
