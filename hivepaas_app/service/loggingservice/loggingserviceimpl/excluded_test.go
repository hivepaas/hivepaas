package loggingserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

func svcWith(id string, driver *swarm.Driver, labels map[string]string) swarm.Service {
	return swarm.Service{ID: id, Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		LogDriver:     driver,
		ContainerSpec: &swarm.ContainerSpec{Labels: labels},
	}}}
}

func TestExcludedAppsReportsEachCause(t *testing.T) {
	jsonFile := appservice.WithLogLabelsOption(&swarm.Driver{Name: "json-file"})
	apps := []*entity.App{
		{ID: "a-good", Name: "good", ServiceID: "s-good"},
		{ID: "a-local", Name: "legacy", ServiceID: "s-local"},
		{ID: "a-old", Name: "unlabelled", ServiceID: "s-old"},
		{ID: "a-clone", Name: "clone", ServiceID: "s-clone"},
		{ID: "a-new", Name: "not deployed yet"},
	}
	services := []swarm.Service{
		svcWith("s-good", jsonFile, map[string]string{appservice.LabelLogAppID: "a-good"}),
		svcWith("s-local", &swarm.Driver{Name: "local"}, nil),
		svcWith("s-old", jsonFile, nil),
		svcWith("s-clone", jsonFile, map[string]string{appservice.LabelLogAppID: "a-source"}),
	}

	got := excludedApps(apps, services)

	assert.Equal(t, []loggingservice.ExcludedApp{
		{AppID: "a-clone", Name: "clone", Driver: "json-file", Reason: loggingservice.ExcludedReasonIdentityMissing},
		{AppID: "a-local", Name: "legacy", Driver: "local", Reason: loggingservice.ExcludedReasonDriverUnreadable},
		{AppID: "a-old", Name: "unlabelled", Driver: "json-file", Reason: loggingservice.ExcludedReasonIdentityMissing},
	}, got)
}

// Nothing to show must be an empty list, not null, for the client.
func TestExcludedAppsIsEmptyNotNil(t *testing.T) {
	assert.NotNil(t, excludedApps(nil, nil))
}
