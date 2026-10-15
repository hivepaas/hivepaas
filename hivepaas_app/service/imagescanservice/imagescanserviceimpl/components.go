package imagescanserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// scanComponents scans what HivePaaS's own parts run, as the daily scan does
// the apps', under the global ignores; and removes the results of the parts
// that do not run any more - a registry switched off. It answers how many it
// scanned, and how many of them failed.
func (s *service) scanComponents(ctx context.Context, ignoreSettings []*entity.Setting, with *imageScanReq) (
	scanned, failed int, err error) {
	ignores := imagescanservice.ComponentIgnores(ignoreSettings)
	running := []string{}
	for _, component := range imagescanservice.Components {
		if err = ctx.Err(); err != nil {
			return scanned, failed, hperrors.Wrap(err)
		}
		images, found, imagesErr := s.componentImages(ctx, component, with.log)
		if !found {
			continue
		}
		running = append(running, component.ID())
		_ = with.log.Add(ctx, tasklog.NewOutFrame("--- HivePaaS / "+component.Name, tasklog.TsNow))
		scanned++
		if _, scanErr := s.scanSubject(ctx, &subject{id: component.ID(), images: images, imagesErr: imagesErr,
			ignores: ignores}, with); scanErr != nil {
			failed++
			_ = with.log.Add(ctx, tasklog.NewWarnFrame("Not scanned: "+hperrors.GetErrorDetail(scanErr, ""),
				tasklog.TsNow))
		}
	}
	return scanned, failed, hperrors.Wrap(s.imageScanRepo.DeleteComponentScansExcept(ctx, s.db, running))
}

// componentImages are the images a part of HivePaaS runs; found is false for
// a part this installation does not run.
func (s *service) componentImages(ctx context.Context, component imagescanservice.Component,
	log *tasklog.Store) (images []runningImage, found bool, err error) {
	switch {
	case component.Service != "":
		svc, err := s.dockerManager.ServiceInspect(ctx, component.Service)
		if errors.Is(err, hperrors.ErrInfraNotFound) {
			return nil, false, nil
		}
		if err != nil {
			return nil, true, hperrors.Wrap(err)
		}
		return s.runningImages(ctx, svc.Service.ID, component, log)

	case component.SystemAppKey != "":
		app, err := s.systemAppService.LoadApp(ctx, s.db, component.SystemAppKey)
		if err != nil {
			return nil, true, hperrors.Wrap(err)
		}
		if app == nil || app.ServiceID == "" {
			return nil, false, nil
		}
		return s.runningImages(ctx, app.ServiceID, component, log)

	case component.Container != "":
		images, err = s.containerOnNodes(ctx, component.Container, log)
		return images, len(images) > 0 || err != nil, err
	}
	return nil, false, nil
}

// runningImages are what a part's service runs. A service with no task
// running - scaled to zero - runs nothing: the part is not running.
func (s *service) runningImages(ctx context.Context, serviceID string, component imagescanservice.Component,
	log *tasklog.Store) ([]runningImage, bool, error) {
	images, err := s.serviceImages(ctx, serviceID, component.Name, log)
	if errors.Is(err, hperrors.ErrActiveContainerNotFound) {
		return nil, false, nil
	}
	return images, true, err
}

// containerOnNodes are the images a container of that name runs, on the ready
// nodes that run one. A node without one, or whose agent does not answer, is
// passed over.
func (s *service) containerOnNodes(ctx context.Context, name string, log *tasklog.Store) ([]runningImage, error) {
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	containers := map[string][]string{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		containers[node.ID] = []string{name}
	}
	var images []runningImage
	nodeIDs := make([]string, 0, len(containers))
	for nodeID := range containers {
		nodeIDs = append(nodeIDs, nodeID)
	}
	slices.Sort(nodeIDs)
	seen := map[string]bool{}
	runningOn := 0
	for _, nodeID := range nodeIDs {
		found, err := s.containerImages(ctx, nodeID, containers[nodeID], tasklog.NewNullStore())
		if err != nil || len(found) == 0 {
			continue
		}
		runningOn++
		for _, img := range found {
			if !seen[img.ID] {
				seen[img.ID] = true
				images = append(images, img)
			}
		}
	}
	if len(images) > 0 {
		_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("%s runs on %d of %d nodes", name, runningOn,
			len(nodeIDs)), tasklog.TsNow))
	}
	return images, nil
}
