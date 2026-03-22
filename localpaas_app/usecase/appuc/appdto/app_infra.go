package appdto

import (
	"github.com/docker/docker/api/types/network"
)

type InfraRefObjects struct {
	Networks map[string]*network.Summary
}
