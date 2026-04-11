package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
)

func ContainerCommandBuild(cmd []string, args []string) string {
	return strings.Join(gofn.Concat(cmd, args), " ")
}

func ContainerCommandApply(contSpec *swarm.ContainerSpec, cmd string) {
	if cmd == "" {
		contSpec.Command = nil
	} else {
		contSpec.Command = gofn.Must(shellutil.CmdSplit(cmd))
	}
}

func CallRetry(
	fn func() error,
	maxRetries int,
	retryInterval time.Duration,
) error {
	retry := 0
	for {
		err := fn()
		if err == nil {
			return nil
		}
		if retry >= maxRetries {
			return apperrors.NewInfra(err)
		}
		if retryInterval > 0 {
			time.Sleep(retryInterval)
		}
		retry++
	}
}

func CallRetry2[T any](
	fn func() (T, error),
	maxRetries int,
	retryInterval time.Duration,
) (T, error) {
	retry := 0
	for {
		v, err := fn()
		if err == nil {
			return v, nil
		}
		if retry >= maxRetries {
			return v, apperrors.NewInfra(err)
		}
		if retryInterval > 0 {
			time.Sleep(retryInterval)
		}
		retry++
	}
}

func CallRetry3[T any, U any](
	fn func() (T, U, error),
	maxRetries int,
	retryInterval time.Duration,
) (T, U, error) {
	retry := 0
	for {
		v1, v2, err := fn()
		if err == nil {
			return v1, v2, nil
		}
		if retry >= maxRetries {
			return v1, v2, apperrors.NewInfra(err)
		}
		if retryInterval > 0 {
			time.Sleep(retryInterval)
		}
		retry++
	}
}
