package commandservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type BuildCommandReq struct {
	Scope             *entity.ObjectScope
	Command           *entity.CommandTemplate
	BuildFinalCommand bool // true to build final command string (not include ENV ref in content)
	RefObjects        *entity.RefObjects
}

type BuildCommandResp struct {
	EnvVars       []*envvarservice.EnvVar
	CommandString string // if req.BuildFinalCommand=true, this is the final command
}
