package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type TaskPeriodicArgs struct {
}

type TaskPeriodicOutput struct {
	Healthcheck *TaskPeriodicHealthcheckOutput `json:"healthcheck,omitempty"`
}

type TaskPeriodicHealthcheckOutput struct {
	REST *TaskPeriodicHealthcheckOutputREST `json:"rest,omitempty"`
	GRPC *TaskPeriodicHealthcheckOutputGRPC `json:"grpc,omitempty"`
}

type TaskPeriodicHealthcheckOutputREST struct {
	ReturnCode int    `json:"returnCode,omitempty"`
	ReturnText string `json:"returnText,omitempty"`
}

type TaskPeriodicHealthcheckOutputGRPC struct {
	ReturnStatus base.HealthcheckGRPCStatus `json:"returnStatus,omitempty"`
}

func (t *Task) ArgsAsPeriodicJob() (*TaskPeriodicArgs, error) {
	return parseTaskArgsAs(t, func() *TaskPeriodicArgs { return &TaskPeriodicArgs{} })
}

func (t *Task) OutputAsPeriodicJob() (*TaskPeriodicOutput, error) {
	return parseTaskOutputAs(t, func() *TaskPeriodicOutput { return &TaskPeriodicOutput{} })
}
