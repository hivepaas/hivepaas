package schedjobdto

import (
	"fmt"
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// SchedJobTriggerReq is an event that runs the job.
type SchedJobTriggerReq struct {
	Event base.SchedJobTriggerEvent `json:"event"`
	// Apps are the apps whose events run the job: none for an app's job, one or
	// more of the env's for an env job.
	Apps []basedto.ObjectIDReq `json:"apps"`
	// Wait holds the deploy until the run ends. pre-deploy only.
	Wait bool `json:"wait"`
}

func triggersToEntity(reqs []*SchedJobTriggerReq) []*entity.SchedJobTrigger {
	triggers := make([]*entity.SchedJobTrigger, 0, len(reqs))
	for _, req := range reqs {
		if req == nil {
			continue
		}
		trigger := &entity.SchedJobTrigger{Event: req.Event, Wait: req.Wait}
		for _, app := range req.Apps {
			trigger.Apps = append(trigger.Apps, entity.ObjectID{ID: app.ID})
		}
		triggers = append(triggers, trigger)
	}
	if len(triggers) == 0 {
		return nil
	}
	return triggers
}

// triggerJobTypes are the job types triggers run: those that run in an app or
// an env. A system job has none.
var triggerJobTypes = []base.SchedJobType{base.SchedJobTypeContainerCommand, base.SchedJobTypeJobSequence}

// validateTriggers checks what a trigger says on its own. Whether its apps are
// the right ones depends on the job's scope, and is checked with it.
func (req *SchedJobBaseReq) validateTriggers(field string) (res []vld.Validator) {
	if len(req.Triggers) == 0 {
		return nil
	}
	res = append(res, basedto.ValidateCond(slices.Contains(triggerJobTypes, req.JobType) &&
		len(req.Triggers) <= base.SchedJobMaxTriggers, field+"triggers")...)
	seen := make(map[string]struct{}, len(req.Triggers))
	for i, trigger := range req.Triggers {
		triggerField := fmt.Sprintf("%striggers[%d]", field, i)
		res = append(res, basedto.ValidateCond(trigger != nil, triggerField)...)
		if trigger == nil {
			continue
		}
		key := triggerKey(trigger)
		_, dup := seen[key]
		seen[key] = struct{}{}
		res = append(res, basedto.ValidateCond(!dup, triggerField)...)
		res = append(res, basedto.ValidateStrIn(&trigger.Event, true, base.AllSchedJobTriggerEvents,
			triggerField+".event")...)
		res = append(res, basedto.ValidateCond(!trigger.Wait || trigger.Event == base.SchedJobTriggerPreDeploy,
			triggerField+".wait")...)
		for j := range trigger.Apps {
			res = append(res, basedto.ValidateObjectIDReq(&trigger.Apps[j], true,
				fmt.Sprintf("%s.apps[%d]", triggerField, j))...)
		}
	}
	return res
}

// triggerKey is what two triggers alike share: their event and their apps.
func triggerKey(trigger *SchedJobTriggerReq) string {
	ids := make([]string, 0, len(trigger.Apps))
	for _, app := range trigger.Apps {
		ids = append(ids, app.ID)
	}
	slices.Sort(ids)
	return string(trigger.Event) + "|" + strings.Join(slices.Compact(ids), ",")
}

type SchedJobTriggerResp struct {
	Event base.SchedJobTriggerEvent  `json:"event"`
	Apps  []*basedto.NamedObjectResp `json:"apps,omitempty"`
	Wait  bool                       `json:"wait,omitempty"`
}

// TransformSchedJobTriggers is a job's triggers with their apps named, from the
// apps refObjects holds. An app gone keeps its ID and no name.
func TransformSchedJobTriggers(
	triggers []*entity.SchedJobTrigger,
	refObjects *entity.RefObjects,
) []*SchedJobTriggerResp {
	resp := make([]*SchedJobTriggerResp, 0, len(triggers))
	for _, trigger := range triggers {
		if trigger == nil {
			continue
		}
		item := &SchedJobTriggerResp{Event: trigger.Event, Wait: trigger.Wait}
		for _, app := range trigger.Apps {
			named := &basedto.NamedObjectResp{ID: app.ID}
			if refObjects != nil {
				if refApp := refObjects.RefApps[app.ID]; refApp != nil {
					named.Name = refApp.Name
				}
			}
			item.Apps = append(item.Apps, named)
		}
		resp = append(resp, item)
	}
	return resp
}
