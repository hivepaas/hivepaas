package taskworkflow_test

import (
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestWorkflowTaskArgsParsing(t *testing.T) {
	wfArgs := &entity.WorkflowArgs{
		CurrentStep: 0,
		Steps: []*entity.WorkflowStep{
			{
				Name: "Step 1",
				Type: base.TaskTypeDummy,
				Args: `{"sleep": "1s"}`,
				Config: &entity.TaskConfig{
					MaxRetry: 3,
				},
			},
			{
				Name: "Step 2",
				Type: base.TaskTypeAppDeploy,
				Args: `{"appId": "app-123"}`,
			},
		},
	}

	task := &entity.Task{
		ID:   "task-wf-1",
		Type: base.TaskTypeWorkflow,
	}

	err := task.SetArgs(wfArgs)
	if err != nil {
		t.Fatalf("SetArgs failed: %v", err)
	}

	parsedWfArgs, err := task.ArgsAsWorkflow()
	if err != nil {
		t.Fatalf("ArgsAsWorkflow failed: %v", err)
	}

	if parsedWfArgs.CurrentStep != 0 {
		t.Fatalf("expected CurrentStep 0, got %d", parsedWfArgs.CurrentStep)
	}
	if len(parsedWfArgs.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(parsedWfArgs.Steps))
	}
	if parsedWfArgs.Steps[0].Name != "Step 1" {
		t.Fatalf("expected Step 1, got %s", parsedWfArgs.Steps[0].Name)
	}
	if parsedWfArgs.Steps[0].Type != base.TaskTypeDummy {
		t.Fatalf("expected TaskTypeDummy, got %s", parsedWfArgs.Steps[0].Type)
	}
	if parsedWfArgs.Steps[0].Config == nil || parsedWfArgs.Steps[0].Config.MaxRetry != 3 {
		t.Fatalf("expected Step 1 MaxRetry 3, got %v", parsedWfArgs.Steps[0].Config)
	}
}
