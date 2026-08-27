package imagebuildserviceimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) imagePush(
	ctx context.Context,
	data *imageBuildData,
) (err error) {
	if data.PushToRegistry.ID == "" {
		return nil
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Start pushing image to registry...",
		tasklog.TsNow))

	regAuth := data.RefObjects.RefSettings[data.PushToRegistry.ID]
	if regAuth == nil {
		return hperrors.NewMissing("Registry auth setting")
	}
	regAuthHeader, err := regAuth.MustAsRegistryAuth().GenerateAuthHeader()
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, tag := range data.Resp.ImageTags {
		if !strings.Contains(tag, "/") { // only push tag containing `/` in it
			continue
		}
		logsReader, err := s.dockerManager.ImagePush(ctx, tag, func(options *client.ImagePushOptions) {
			options.RegistryAuth = regAuthHeader
		})
		if err != nil {
			return hperrors.Wrap(err)
		}

		logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
		for msgs := range logsChan {
			for _, msg := range msgs {
				frameCreator := tasklog.NewOutFrame
				if msg.Error != nil {
					err = errors.Join(err, msg.Error)
					frameCreator = tasklog.NewErrFrame
				}
				if msg.String() != "" {
					_ = data.LogStore.Add(ctx, frameCreator(msg.String(), tasklog.TsNow))
				}
			}
		}
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}
