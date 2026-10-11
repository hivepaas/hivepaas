package hperrors

// Errors for image scanning
var (
	// ErrImageScanOff refuses a scan while image scanning is switched off.
	ErrImageScanOff = NewErr(ErrPreconditionFailed, "ERR_IMAGE_SCAN_OFF")
	// ErrImageScanNotReady is a scan the scanner cannot take yet: its apps are
	// being made, or are not running on the node the image is on.
	ErrImageScanNotReady = NewErr(ErrUnavailable, "ERR_IMAGE_SCAN_NOT_READY")
	// ErrImageScanFailed is a scan Trivy did not finish; the detail is its own.
	ErrImageScanFailed = NewErr(ErrActionFailed, "ERR_IMAGE_SCAN_FAILED")
)
