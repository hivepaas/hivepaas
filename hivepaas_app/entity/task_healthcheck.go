package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type TaskHealthcheckArgs struct {
}

type TaskHealthcheckOutput struct {
	REST *TaskHealthcheckOutputREST `json:"rest,omitempty"`
	GRPC *TaskHealthcheckOutputGRPC `json:"grpc,omitempty"`
}

type TaskHealthcheckOutputREST struct {
	ReturnCode int    `json:"returnCode,omitempty"`
	ReturnText string `json:"returnText,omitempty"`
}

type TaskHealthcheckOutputGRPC struct {
	ReturnStatus base.HealthcheckGRPCStatus `json:"returnStatus,omitempty"`
}

func (t *Task) ArgsAsHealthcheck() (*TaskHealthcheckArgs, error) {
	return parseTaskArgsAs(t, func() *TaskHealthcheckArgs { return &TaskHealthcheckArgs{} })
}

func (t *Task) OutputAsHealthcheck() (*TaskHealthcheckOutput, error) {
	return parseTaskOutputAs(t, func() *TaskHealthcheckOutput { return &TaskHealthcheckOutput{} })
}
