package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// SchedJobTrigger is an event that runs a scheduled job.
type SchedJobTrigger struct {
	Event base.SchedJobTriggerEvent `json:"event"`
	// Apps are the apps whose events run the job: empty for an app's job, which
	// listens to its own app; one or more apps of the env for an env job.
	Apps []ObjectID `json:"apps,omitempty"`
	// Wait holds the deploy until the run ends. pre-deploy only.
	Wait bool `json:"wait,omitempty"`
}

// ListensTo says whether the job runs when event happens to the app appID, and
// whether a trigger that matches waits. ownApp is true when appID is the app the
// job belongs to: a trigger naming no app listens to that one.
func (s *SchedJob) ListensTo(event base.SchedJobTriggerEvent, appID string, ownApp bool) (listens, wait bool) {
	if s == nil {
		return false, false
	}
	for _, trigger := range s.Triggers {
		if trigger == nil || trigger.Event != event || !trigger.names(appID, ownApp) {
			continue
		}
		listens = true
		wait = wait || trigger.Wait
	}
	return listens, wait
}

func (t *SchedJobTrigger) names(appID string, ownApp bool) bool {
	if len(t.Apps) == 0 {
		return ownApp
	}
	for _, app := range t.Apps {
		if app.ID == appID {
			return true
		}
	}
	return false
}

// TriggerAppIDs is the apps the job's triggers name, each once.
func (s *SchedJob) TriggerAppIDs() []string {
	if s == nil {
		return nil
	}
	var ids []string
	seen := map[string]struct{}{}
	for _, trigger := range s.Triggers {
		if trigger == nil {
			continue
		}
		for _, app := range trigger.Apps {
			if _, dup := seen[app.ID]; dup || app.ID == "" {
				continue
			}
			seen[app.ID] = struct{}{}
			ids = append(ids, app.ID)
		}
	}
	return ids
}

// SchedJobTriggerCause is what fired a scheduled job's run.
type SchedJobTriggerCause struct {
	Event        base.SchedJobTriggerEvent `json:"event"`
	AppID        string                    `json:"appId"`
	DeploymentID string                    `json:"deploymentId,omitempty"`
}

// TaskSchedJobExecArgs are the args of a scheduled job's run: what fired it,
// when a trigger did.
type TaskSchedJobExecArgs struct {
	Trigger *SchedJobTriggerCause `json:"trigger,omitempty"`
}

// ArgsAsSchedJobExec is a run's args; nil for a run no trigger fired.
func (t *Task) ArgsAsSchedJobExec() (*TaskSchedJobExecArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSchedJobExecArgs { return &TaskSchedJobExecArgs{} })
}
