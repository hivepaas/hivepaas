package networkserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

// lingeringDocker has a network of the name for as many inspections as given,
// then none - what a deleted project's network does while its containers stop:
// it is there, yet a service cannot use it nor a network take its name.
type lingeringDocker struct {
	docker.Manager
	inspections int
	lingerFor   int
	labels      map[string]string
	created     []string
}

func (d *lingeringDocker) NetworkInspect(
	_ context.Context, name string, _ ...docker.NetworkInspectOption,
) (*client.NetworkInspectResult, error) {
	d.inspections++
	if len(d.created) == 0 && d.inspections > d.lingerFor {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.NetworkInspectResult{Network: network.Inspect{
		Network: network.Network{ID: "net-" + name, Name: name, Driver: "overlay", Labels: d.labels},
	}}, nil
}

func (d *lingeringDocker) NetworkCreate(
	_ context.Context, name string, options ...docker.NetworkCreateOption,
) (*client.NetworkCreateResult, error) {
	opts := client.NetworkCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	d.labels = opts.Labels
	d.created = append(d.created, name)
	return &client.NetworkCreateResult{ID: "net-" + name}, nil
}

type noNetworkSettings struct {
	repository.SettingRepo
	upserted []*entity.Setting
}

func (r *noNetworkSettings) GetByName(context.Context, database.IDB, *entity.ObjectScope, base.SettingType,
	string, bool, ...bunex.SelectQueryOption) (*entity.Setting, error) {
	return nil, hperrors.Wrap(hperrors.ErrNotFound)
}

func (r *noNetworkSettings) Upsert(_ context.Context, _ database.IDB, setting *entity.Setting,
	_, _ []string, _ ...bunex.InsertQueryOption) error {
	r.upserted = append(r.upserted, setting)
	return nil
}

// A project deleted and made again at once - restored from its spec, a Compose
// file applied again - finds the network of its env's name still there for the
// seconds its old containers take to stop. Taken as its own, its first app
// failed: "network ... not found". It is waited out, and the network made anew.
func TestAnEnvNetworkLeftByADeletedProjectIsWaitedOutAndMadeAgain(t *testing.T) {
	networkLingerPoll = time.Millisecond
	dockerMgr := &lingeringDocker{lingerFor: 3, labels: map[string]string{docker.StackLabelNamespace: "shop"}}
	settings := &noNetworkSettings{}
	s := &service{dockerManager: dockerMgr, settingRepo: settings}

	setting, inspect, err := s.GetOrCreateProjectNetwork(context.Background(), nil,
		&entity.Project{ID: "P2", Key: "shop"}, "prod")

	assert.NoError(t, err)
	assert.Equal(t, []string{"shop_prod_net"}, dockerMgr.created, "made again, once the old one was gone")
	assert.Equal(t, "P2", dockerMgr.labels[labelProjectID], "labeled with the project it is made for")
	assert.Equal(t, "shop_prod_net", inspect.Name)
	assert.Equal(t, "P2", setting.ObjectID)
}

// The project's own network - labeled with its id - is used as it is, at once.
func TestAProjectsOwnEnvNetworkIsUsedAtOnce(t *testing.T) {
	networkLingerPoll = time.Millisecond
	dockerMgr := &lingeringDocker{lingerFor: 1000, labels: map[string]string{labelProjectID: "P2"}}
	s := &service{dockerManager: dockerMgr, settingRepo: &noNetworkSettings{}}

	_, _, err := s.GetOrCreateProjectNetwork(context.Background(), nil, &entity.Project{ID: "P2", Key: "shop"}, "prod")

	assert.NoError(t, err)
	assert.Empty(t, dockerMgr.created)
	assert.Equal(t, 1, dockerMgr.inspections)
}
