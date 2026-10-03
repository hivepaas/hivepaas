package config

import "time"

type Tasks struct {
	Queue    TaskQueue    `toml:"queue"`
	Periodic Periodic     `toml:"periodic"`
	Triggers TaskTriggers `toml:"triggers"`
}

type TaskQueue struct {
	Concurrency        int           `toml:"concurrency" env:"HP_TASKS_QUEUE_CONCURRENCY" default:"10"`
	TaskCheckInterval  time.Duration `toml:"task_check_interval" env:"HP_TASKS_QUEUE_TASK_CHECK_INTERVAL" default:"10m"`
	TaskCreateInterval time.Duration `toml:"task_create_interval" env:"HP_TASKS_QUEUE_TASK_CREATE_INTERVAL" default:"10m"`
}

type Periodic struct {
	BaseInterval time.Duration `toml:"base_interval" env:"HP_TASKS_PERIODIC_BASE_INTERVAL" default:"2s"`
	BatchSize    int           `toml:"batch_size" env:"HP_TASKS_PERIODIC_BATCH_SIZE" default:"100"`
}

// TaskTriggers is how a deploy waits for the pre-deploy jobs that hold it.
type TaskTriggers struct {
	// WaitPollInterval is how often the deploy reads the status of the runs.
	WaitPollInterval time.Duration `toml:"wait_poll_interval" env:"HP_TASKS_TRIGGERS_WAIT_POLL_INTERVAL" default:"3s"`
	// WaitTimeout is how long it waits for a job without a timeout of its own,
	// and for a job sequence, whose timeout is each step's.
	WaitTimeout time.Duration `toml:"wait_timeout" env:"HP_TASKS_TRIGGERS_WAIT_TIMEOUT" default:"30m"`
}
