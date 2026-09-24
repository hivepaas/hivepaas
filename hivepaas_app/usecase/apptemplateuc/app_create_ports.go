package apptemplateuc

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/network"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// checkPublishedPorts refuses a request whose apps cannot have the ports they
// ask for, before anything is created.
//
// A published port is an address on every node of the cluster, so two apps
// cannot share one. Docker refuses the second itself, but only while the
// service is being created: a request creating an app and its database would
// have created the database by then, and the message would be about swarm
// rather than about the port. This runs the same check the app's network
// settings run, plus one they cannot: two apps created together must not ask
// for the same port either.
func (uc *UC) checkPublishedPorts(ctx context.Context, apps []*appToProvision) error {
	claimed := map[clusterservice.PortRef]string{}
	wanted := make([]clusterservice.PortRef, 0, len(apps))
	for _, target := range apps {
		for _, port := range specmodel.PublishedPorts(target.result.Doc) {
			ref := clusterservice.PortRef{Published: port.Published, Protocol: port.Protocol}
			if ref.Protocol == "" {
				ref.Protocol = network.TCP
			}
			if by, taken := claimed[ref]; taken {
				return hperrors.Wrap(hperrors.ErrPortInUse).
					WithParam("Port", ref.Published).
					WithParam("Protocol", string(ref.Protocol)).
					WithParam("PublishedBy", by).
					WithExtraDetail("%s and %s are created together and ask for the same port %s",
						by, target.name, describePort(ref))
			}
			claimed[ref] = target.name
			wanted = append(wanted, ref)
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	return hperrors.Wrap(uc.clusterService.VerifyPortsAvailable(ctx, wanted, nil))
}

func describePort(ref clusterservice.PortRef) string {
	return fmt.Sprintf("%d/%s", ref.Published, strings.ToLower(string(ref.Protocol)))
}
