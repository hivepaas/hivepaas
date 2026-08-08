package config

import "time"

type Tasks struct {
	Queue    TaskQueue `toml:"queue"`
	Periodic Periodic  `toml:"periodic"`
}

type TaskQueue struct {
	Concurrency        int           `toml:"concurrency" env:"HP_TASKS_QUEUE_CONCURRENCY" default:"10"`
	TaskCheckInterval  time.Duration `toml:"task_check_interval" env:"HP_TASKS_QUEUE_TASK_CHECK_INTERVAL" default:"10m"`
	TaskCreateInterval time.Duration `toml:"task_create_interval" env:"HP_TASKS_QUEUE_TASK_CREATE_INTERVAL" default:"10m"`
}

type Periodic struct {
	BaseInterval time.Duration `toml:"base_interval" env:"HP_TASKS_PERIODIC_BASE_INTERVAL" default:"15s"`
	BatchSize    int           `toml:"batch_size" env:"HP_TASKS_PERIODIC_BATCH_SIZE" default:"100"`
}
