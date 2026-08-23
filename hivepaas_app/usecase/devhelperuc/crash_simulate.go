package devhelperuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) SimulateCrash(
	ctx context.Context,
	req *devhelperdto.SimulateCrashReq,
) (*devhelperdto.SimulateCrashResp, error) {
	panic("app crashes")
}
