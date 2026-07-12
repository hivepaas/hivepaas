package imagebuildserviceimpl

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type imageBuildData struct {
	*imagebuildservice.ImageBuildReq
	Resp *imagebuildservice.ImageBuildResp
}

func (s *service) ImageBuild(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (resp *imagebuildservice.ImageBuildResp, err error) {
	resp = &imagebuildservice.ImageBuildResp{}
	data := &imageBuildData{
		ImageBuildReq: req,
		Resp:          resp,
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
		}
	}()

	err = s.doImageBuild(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.doImagePush(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, err
}

//nolint:gocognit
func (s *service) doImageBuild(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (err error) {
	repoSource := data.RepoSource
	buildSetting := data.ImageBuildSettings

	s.addStepStartLog(ctx, data, "Start building image...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	// TODO: check dockerfile existence
	dockerfile := gofn.Coalesce(repoSource.DockerfilePath, "Dockerfile")

	imageTags, err := s.calcBuildImageTags(repoSource.ImageTags, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Resp.ImageTags = imageTags

	envVars, err := s.calcBuildEnvVars(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	authConfigs, err := s.calcBuildRegistryAuths(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create tar archive for the source code
	tar, err := archive.TarWithOptions(data.CheckoutDir, &archive.TarOptions{})
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer tar.Close()

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return apperrors.Wrap(err)
	}

	// Build the image
	resp, err := s.dockerManager.ImageBuild(ctx, tar, func(opts *client.ImageBuildOptions) {
		opts.Version = build.BuilderV1
		opts.BuildID = data.BuildID
		opts.Dockerfile = dockerfile
		opts.Tags = imageTags
		opts.BuildArgs = envVars
		opts.AuthConfigs = authConfigs

		if buildSetting != nil {
			opts.NoCache = buildSetting.NoCache || data.NoCache
			opts.SuppressOutput = buildSetting.NoVerbose
			res := buildSetting.Resources
			if res.CPUs > 0 {
				opts.CPUPeriod, opts.CPUQuota = res.CPUsAsPeriodAndQuota()
			}
			if res.Mem > 0 {
				opts.Memory = res.Mem.Bytes()
			}
			if res.MemSwap > 0 {
				opts.MemorySwap = res.MemSwap.Bytes()
			}
			if res.ShmSize > 0 {
				opts.ShmSize = res.ShmSize.Bytes()
			}
		}
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, resp.Body, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := tasklog.NewDebugFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = tasklog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.AddRedacted(ctx, frameCreator(msg.String(), tasklog.TsNow))
			}
		}
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) doImagePush(
	ctx context.Context,
	data *imageBuildData,
) (err error) {
	repoSource := data.RepoSource
	if repoSource.PushToRegistry.ID == "" {
		return nil
	}

	s.addStepStartLog(ctx, data, "Start pushing image to registry...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	var regAuthHeader string
	if repoSource.PushToRegistry.ID != "" {
		regAuth := data.RefObjects.RefSettings[repoSource.PushToRegistry.ID]
		regAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	for _, tag := range data.Resp.ImageTags {
		if !strings.Contains(tag, "/") { // only push tag containing `/` in it
			continue
		}
		logsReader, err := s.dockerManager.ImagePush(ctx, tag, func(options *client.ImagePushOptions) {
			options.RegistryAuth = regAuthHeader
		})
		if err != nil {
			return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) addStepStartLog(
	ctx context.Context,
	data *imageBuildData,
	msg string,
) {
	_ = data.LogStore.Add(ctx,
		tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
		tasklog.NewOutFrame(msg, tasklog.TsNow))
}

func (s *service) addStepEndLog(
	ctx context.Context,
	data *imageBuildData,
	start time.Time,
	err error,
) {
	duration := timeutil.NowUTC().Sub(start).Truncate(time.Millisecond)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String()+
			" with error: "+err.Error(), tasklog.TsNow))
	} else {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String(),
			tasklog.TsNow))
	}
}
