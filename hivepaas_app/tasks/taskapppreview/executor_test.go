package taskapppreview_test

import (
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestAppPreviewTaskArgsAndOutputParsing(t *testing.T) {
	previewArgs := &entity.TaskAppPreviewArgs{
		ParentApp:       entity.ObjectID{ID: "app-parent-123"},
		RepoRef:         "refs/pull/42/head",
		CustomSubdomain: "custom-feat",
		NoStart:         true,
		CloneDBApps:     true,
		Trigger: &entity.AppDeploymentTrigger{
			Source:   base.DeploymentTriggerSourceUser,
			SourceID: "user-1",
		},
	}

	task := &entity.Task{
		ID:   "task-preview-1",
		Type: base.TaskTypeAppPreview,
	}

	err := task.SetArgs(previewArgs)
	if err != nil {
		t.Fatalf("SetArgs failed: %v", err)
	}

	parsedArgs, err := task.ArgsAsAppPreview()
	if err != nil {
		t.Fatalf("ArgsAsAppPreview failed: %v", err)
	}

	if parsedArgs.ParentApp.ID != "app-parent-123" {
		t.Fatalf("expected ParentApp ID app-parent-123, got %s", parsedArgs.ParentApp.ID)
	}
	if parsedArgs.RepoRef != "refs/pull/42/head" {
		t.Fatalf("expected RepoRef refs/pull/42/head, got %s", parsedArgs.RepoRef)
	}
	if parsedArgs.CustomSubdomain != "custom-feat" {
		t.Fatalf("expected CustomSubdomain custom-feat, got %s", parsedArgs.CustomSubdomain)
	}
	if !parsedArgs.NoStart {
		t.Fatalf("expected NoStart true, got false")
	}
	if !parsedArgs.CloneDBApps {
		t.Fatalf("expected CloneDBApps true, got false")
	}
	if parsedArgs.Trigger == nil || parsedArgs.Trigger.Source != base.DeploymentTriggerSourceUser {
		t.Fatalf("expected Trigger Source user, got %v", parsedArgs.Trigger)
	}

	previewOutput := &entity.TaskAppPreviewOutput{
		PreviewApp: entity.ObjectID{ID: "app-preview-123"},
		Deployment: entity.ObjectID{ID: "deploy-123"},
		ClonedDBApps: entity.ObjectIDSlice{
			{ID: "db-clone-1"},
		},
	}

	err = task.SetOutput(previewOutput)
	if err != nil {
		t.Fatalf("SetOutput failed: %v", err)
	}

	parsedOutput, err := task.OutputAsAppPreview()
	if err != nil {
		t.Fatalf("OutputAsAppPreview failed: %v", err)
	}

	if parsedOutput.PreviewApp.ID != "app-preview-123" {
		t.Fatalf("expected PreviewApp ID app-preview-123, got %s", parsedOutput.PreviewApp.ID)
	}
	if parsedOutput.Deployment.ID != "deploy-123" {
		t.Fatalf("expected Deployment ID deploy-123, got %s", parsedOutput.Deployment.ID)
	}
	if len(parsedOutput.ClonedDBApps) != 1 || parsedOutput.ClonedDBApps[0].ID != "db-clone-1" {
		t.Fatalf("expected 1 cloned db app with ID db-clone-1, got %v", parsedOutput.ClonedDBApps)
	}
}
