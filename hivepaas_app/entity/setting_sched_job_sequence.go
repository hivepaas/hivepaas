package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// SchedJobSequence is the jobs a job sequence runs, in order.
type SchedJobSequence struct {
	Mode      base.SchedJobSeqMode      `json:"mode"`
	OnFailure base.SchedJobSeqOnFailure `json:"onFailure"`
	Steps     []*SchedJobSequenceStep   `json:"steps"`
}

// SchedJobSequenceStep is one scheduled job a sequence runs, with an optional
// label for its logs and results. A job may be a step more than once.
type SchedJobSequenceStep struct {
	Job  ObjectID `json:"job"`
	Name string   `json:"name,omitempty"`
}

// MemberIDs is the sequence's jobs, each once, in the order they first appear.
func (s *SchedJobSequence) MemberIDs() []string {
	if s == nil {
		return nil
	}
	ids := make([]string, 0, len(s.Steps))
	seen := make(map[string]struct{}, len(s.Steps))
	for _, step := range s.Steps {
		if step == nil || step.Job.ID == "" {
			continue
		}
		if _, dup := seen[step.Job.ID]; dup {
			continue
		}
		seen[step.Job.ID] = struct{}{}
		ids = append(ids, step.Job.ID)
	}
	return ids
}

// StopsOnFailure says whether a step failing for good ends the run.
func (s *SchedJobSequence) StopsOnFailure() bool {
	return s == nil || s.OnFailure != base.SchedJobSeqOnFailureContinue
}
