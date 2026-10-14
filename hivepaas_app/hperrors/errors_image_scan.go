package hperrors

// Errors for image scanning
var (
	// ErrImageScanOff refuses a scan while image scanning is switched off.
	ErrImageScanOff = NewErr(ErrPreconditionFailed, "ERR_IMAGE_SCAN_OFF")
	// ErrImageScanNotReady is a scan that cannot run now: its volume is gone,
	// or the node it would run on cannot be reached.
	ErrImageScanNotReady = NewErr(ErrUnavailable, "ERR_IMAGE_SCAN_NOT_READY")
	// ErrImageScanFailed is a scan Trivy did not finish; the detail is its own.
	ErrImageScanFailed = NewErr(ErrActionFailed, "ERR_IMAGE_SCAN_FAILED")
	// ErrImageScanSettingsInvalid refuses settings scans could not run by.
	ErrImageScanSettingsInvalid = NewErr(ErrValueInvalid, "ERR_IMAGE_SCAN_SETTINGS_INVALID")
)
