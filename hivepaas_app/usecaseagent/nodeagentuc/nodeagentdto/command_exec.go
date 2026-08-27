package nodeagentdto

import "github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"

type ExecCommandReq struct {
	*nodeexecservice.CommandExecOpts
}

type ExecCommandResp struct {
	ExitCode int32
}
