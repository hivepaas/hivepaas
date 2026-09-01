package kopia

import (
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

type Client struct {
	storage     *backupmodel.Storage
	commandExec backupmodel.CommandExecutor
}

func NewClient(
	storage *backupmodel.Storage,
	commandExec backupmodel.CommandExecutor,
) *Client {
	return &Client{storage: storage, commandExec: commandExec}
}

func (c *Client) Name() backupmodel.EngineType {
	return backupmodel.EngineTypeKopia
}
