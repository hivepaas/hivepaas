package base

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusDisabled ProjectStatus = "disabled"
	ProjectStatusDeleting ProjectStatus = "deleting"
)

var (
	AllProjectStatuses = []ProjectStatus{ProjectStatusActive, ProjectStatusDisabled,
		ProjectStatusDeleting}
)
