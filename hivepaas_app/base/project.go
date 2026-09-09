package base

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusDisabled ProjectStatus = "disabled"
	ProjectStatusDeleting ProjectStatus = "deleting"
	ProjectStatusMissing  ProjectStatus = "missing" // NOTE: this is not used in DB
)

var (
	AllProjectStatuses = []ProjectStatus{ProjectStatusActive, ProjectStatusDisabled,
		ProjectStatusDeleting, ProjectStatusMissing}
)
