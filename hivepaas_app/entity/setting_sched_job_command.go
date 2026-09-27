package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type SchedJobCommandOutput struct {
	Enabled    bool                             `json:"enabled"`
	SaveToFile *SchedJobCommandOutputSaveToFile `json:"saveToFile,omitempty"`
	PipeToApp  *SchedJobCommandOutputPipeToApp  `json:"pipeToApp,omitempty"`
}

type SchedJobCommandOutputSaveToFile struct {
	FileName          string                           `json:"fileName"`
	FilePath          string                           `json:"filePath"`
	FileKind          base.FileKind                    `json:"fileKind"`
	Storage           SchedJobCommandOutputFileStorage `json:"storage"`
	CompressionFormat base.FileCompressionFormat       `json:"compressionFormat"`
	EncryptionFormat  base.FileEncryptionFormat        `json:"encryptionFormat"`
	EncryptionSecret  EncryptedField                   `json:"encryptionSecret"`
}

type SchedJobCommandOutputFileStorage struct {
	ID     string `json:"id"`
	Bucket string `json:"bucket,omitempty"`
}

type SchedJobCommandOutputPipeToApp struct {
	TargetApp ObjectID         `json:"targetApp"`
	Command   *CommandTemplate `json:"command"`
}
