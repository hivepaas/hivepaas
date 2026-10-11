package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// TaskImageScanArgs is a scan of an app's images. The task's object is the app.
type TaskImageScanArgs struct {
	AppID   string                `json:"appId"`
	Trigger base.ImageScanTrigger `json:"trigger"`
	// Force scans an image again though it was scanned within the hour.
	Force bool `json:"force,omitempty"`
}

// TaskImageScanOutput is what the scan found, in counts.
type TaskImageScanOutput struct {
	Images []*TaskImageScanImage `json:"images,omitempty"`
	Error  string                `json:"error,omitempty"`
}

type TaskImageScanImage struct {
	ImageID  string `json:"imageId"`
	ImageRef string `json:"imageRef,omitempty"`
	// Skipped says the image was scanned within the hour, and was not again.
	Skipped  bool   `json:"skipped,omitempty"`
	Critical int    `json:"critical"`
	High     int    `json:"high"`
	Fixable  int    `json:"fixable"`
	Error    string `json:"error,omitempty"`
}

func (t *Task) ArgsAsImageScan() (*TaskImageScanArgs, error) {
	return parseTaskArgsAs(t, func() *TaskImageScanArgs { return &TaskImageScanArgs{} })
}

func (t *Task) OutputAsImageScan() (*TaskImageScanOutput, error) {
	return parseTaskOutputAs(t, func() *TaskImageScanOutput { return &TaskImageScanOutput{} })
}
