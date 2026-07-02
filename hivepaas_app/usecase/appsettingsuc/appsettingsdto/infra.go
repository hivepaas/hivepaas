package appsettingsdto

import (
	"github.com/moby/moby/api/types/network"
)

type InfraRefObjects struct {
	Networks map[string]*network.Summary
}
