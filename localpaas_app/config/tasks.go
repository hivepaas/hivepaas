package config

import "time"

type Tasks struct {
	Queue       TaskQueue   `toml:"queue"`
	Healthcheck Healthcheck `toml:"healthcheck"`
}

type TaskQueue struct {
	Concurrency        int           `toml:"concurrency" env:"LP_TASKS_QUEUE_CONCURRENCY" default:"5"`
	TaskCheckInterval  time.Duration `toml:"task_check_interval" env:"LP_TASKS_QUEUE_TASK_CHECK_INTERVAL" default:"10m"`
	TaskCreateInterval time.Duration `toml:"task_create_interval" env:"LP_TASKS_QUEUE_TASK_CREATE_INTERVAL" default:"10m"`
}

type Healthcheck struct {
	BaseInterval time.Duration `toml:"base_interval" env:"LP_TASKS_HEALTHCHECK_BASE_INTERVAL" default:"15s"`
}
