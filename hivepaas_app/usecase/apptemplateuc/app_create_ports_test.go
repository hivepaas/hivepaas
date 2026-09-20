package apptemplateuc

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

// portTemplateYAML publishes one UDP port, the way a VPN or a DNS server does.
const portTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: vpn
  title: VPN
  tagline: Test VPN
  description: Test.
  categories: [webapps/networking]
  icon: icons/vpn.svg
  requires: {versionCode: v000001}
parameters:
  - {name: vpnPort, title: VPN port, type: int, default: 51820}
versions:
  - {name: "1", release: "1.0", default: true, image: "wg:1.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    networks:
      endpointSpec:
        ports:
          - target: "${{ params.vpnPort }}"
            published: "${{ params.vpnPort }}"
            protocol: udp
            publishMode: host
`

type fakeClusterService struct {
	clusterservice.Service
	taken  []clusterservice.PortRef
	asked  []clusterservice.PortRef
	called bool
}

func (f *fakeClusterService) VerifyPortsAvailable(
	_ context.Context, ports []clusterservice.PortRef, _ []string,
) error {
	f.called = true
	f.asked = ports
	for _, port := range ports {
		for _, busy := range f.taken {
			if port == busy {
				return hperrors.Wrap(hperrors.ErrPortInUse).WithParam("Port", port.Published)
			}
		}
	}
	return nil
}

func portApps(t *testing.T, names ...string) []*appToProvision {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(portTemplateYAML))
	assert.NoError(t, err)
	apps := make([]*appToProvision, 0, len(names))
	for _, name := range names {
		result, renderErr := templaterender.Render(&templaterender.Request{Template: tmpl})
		assert.NoError(t, renderErr)
		apps = append(apps, &appToProvision{
			id:   name,
			name: name,
			rendered: &apptemplateservice.RenderResp{
				TemplateResp: apptemplateservice.TemplateResp{Template: tmpl},
				Result:       result,
			},
		})
	}
	return apps
}

func TestCheckPublishedPortsAsksTheClusterForEveryPort(t *testing.T) {
	cluster := &fakeClusterService{}
	uc := &UC{clusterService: cluster}

	assert.NoError(t, uc.checkPublishedPorts(context.Background(), portApps(t, "vpn")))
	assert.Equal(t, []clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}, cluster.asked)
}

func TestCheckPublishedPortsRefusesAPortTheClusterAlreadyHas(t *testing.T) {
	cluster := &fakeClusterService{taken: []clusterservice.PortRef{{Published: 51820, Protocol: network.UDP}}}
	uc := &UC{clusterService: cluster}

	err := uc.checkPublishedPorts(context.Background(), portApps(t, "vpn"))

	assert.ErrorIs(t, err, hperrors.ErrPortInUse)
}

// Two apps of one request are created together, so one cannot be given a port
// the other is taking - and nothing outside the request knows that yet.
func TestCheckPublishedPortsRefusesTwoAppsOfOneRequestAskingForTheSamePort(t *testing.T) {
	cluster := &fakeClusterService{}
	uc := &UC{clusterService: cluster}

	err := uc.checkPublishedPorts(context.Background(), portApps(t, "vpn-a", "vpn-b"))

	assert.ErrorIs(t, err, hperrors.ErrPortInUse)
	var hpErr hperrors.HPError
	assert.ErrorAs(t, err, &hpErr)
	assert.Contains(t, hpErr.Build("en").Detail, "are created together and ask for the same port 51820/udp")
	assert.False(t, cluster.called, "the request refuses itself before the cluster is asked")
}

func TestCheckPublishedPortsAsksNothingOfATemplateThatPublishesNone(t *testing.T) {
	uc, fakes := newCreateTest(t)
	cluster := &fakeClusterService{}
	uc.clusterService = cluster

	apps := planApps(testCreateReq(), fakes.templates.resp)

	assert.NoError(t, uc.checkPublishedPorts(context.Background(), apps))
	assert.False(t, cluster.called)
}
