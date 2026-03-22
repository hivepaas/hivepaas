package base

import "time"

type DeploymentStatus string

const (
	DeploymentStatusNotStarted DeploymentStatus = "not-started"
	DeploymentStatusInProgress DeploymentStatus = "in-progress"
	DeploymentStatusCanceled   DeploymentStatus = "canceled"
	DeploymentStatusDone       DeploymentStatus = "done"
	DeploymentStatusFailed     DeploymentStatus = "failed"
)

var (
	AllDeploymentStatuses = []DeploymentStatus{DeploymentStatusNotStarted, DeploymentStatusInProgress,
		DeploymentStatusCanceled, DeploymentStatusDone, DeploymentStatusFailed}
)

const (
	// TODO: make this configurable
	DeploymentTimeoutDefault = 60 * time.Minute
)

type DeploymentTriggerSource string

const (
	DeploymentTriggerSourceUser        DeploymentTriggerSource = "user"
	DeploymentTriggerSourceRepoWebhook DeploymentTriggerSource = "repo-webhook"
	DeploymentTriggerSourceAPIWebhook  DeploymentTriggerSource = "api-webhook"
)

var (
	AllDeploymentTriggerSources = []DeploymentTriggerSource{DeploymentTriggerSourceUser,
		DeploymentTriggerSourceRepoWebhook, DeploymentTriggerSourceAPIWebhook}
)
