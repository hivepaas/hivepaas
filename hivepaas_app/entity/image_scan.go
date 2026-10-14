package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ImageScan is the last scan of an image, by its ID: the config digest, the
// same on every node of a platform. Two apps on the same image share it.
type ImageScan struct {
	ImageID string `bun:",pk" json:"imageId"`
	// ImageRef is the reference the image was last seen by, for display.
	ImageRef string `json:"imageRef"`
	Platform string `bun:",nullzero" json:"platform,omitempty"`
	OSFamily string `bun:"os_family,nullzero" json:"osFamily,omitempty"`
	OSName   string `bun:"os_name,nullzero" json:"osName,omitempty"`
	// OSEOSL says the OS is past its support: its packages get no fixes, so
	// what the scan does not find may still be there.
	OSEOSL bool `bun:"os_eosl" json:"osEosl,omitempty"`

	Status base.ImageScanStatus `json:"status"`
	// Source is where the scan read the image: on its node, or from its
	// registry.
	Source       base.ImageScanSource `bun:",nullzero" json:"source,omitempty"`
	Error        string               `bun:",nullzero" json:"error,omitempty"`
	TrivyVersion string               `bun:",nullzero" json:"trivyVersion,omitempty"`
	// DBUpdatedAt is when the vulnerability DB the scan was matched against
	// was built.
	DBUpdatedAt time.Time `bun:"db_updated_at,nullzero" json:"dbUpdatedAt,omitzero"`
	ScannedAt   time.Time `json:"scannedAt"`
	UpdatedAt   time.Time `bun:",default:current_timestamp" json:"updatedAt"`

	Findings []*ImageScanFinding `bun:"rel:has-many,join:image_id=image_id" json:"-"`
}

// ImageScanFinding is one vulnerability of one package of an image. A scan
// replaces its image's findings whole.
type ImageScanFinding struct {
	ImageID          string            `bun:",pk" json:"imageId"`
	VulnID           string            `bun:",pk" json:"vulnId"`
	PkgName          string            `bun:",pk" json:"pkgName"`
	PkgType          string            `json:"pkgType"`
	PkgPath          string            `bun:",pk" json:"pkgPath,omitempty"`
	InstalledVersion string            `bun:",pk" json:"installedVersion"`
	FixedVersion     string            `bun:",nullzero" json:"fixedVersion,omitempty"`
	Severity         base.VulnSeverity `json:"severity"`
	Title            string            `bun:",nullzero" json:"title,omitempty"`
	URL              string            `bun:"url,nullzero" json:"url,omitempty"`
	PublishedAt      time.Time         `bun:",nullzero" json:"publishedAt,omitzero"`
}

// Fixable says a version without the vulnerability exists.
func (f *ImageScanFinding) Fixable() bool {
	return f.FixedVersion != ""
}

// Key is the finding as one string, among an image's.
func (f *ImageScanFinding) Key() string {
	return f.VulnID + "|" + f.PkgName + "|" + f.PkgPath + "|" + f.InstalledVersion
}

// AppImageScan is what an app's tasks ran when it was last scanned: the
// images, or why none could be scanned.
type AppImageScan struct {
	AppID    string   `bun:",pk" json:"appId"`
	ImageIDs []string `bun:"image_ids,array" json:"imageIds"`
	// Error is why the app's images could not be found or scanned; each image's
	// own failure is on its scan.
	Error     string    `bun:",nullzero" json:"error,omitempty"`
	ScannedAt time.Time `json:"scannedAt"`
	// Reported are the keys of the fixable CRITICAL and HIGH findings the daily
	// scan last reported for the app, after ignores: the next run tells only
	// what is new against them.
	Reported []string `bun:",array" json:"reported,omitempty"`
}
