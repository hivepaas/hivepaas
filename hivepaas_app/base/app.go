package base

type AppStatus string

const (
	AppStatusActive   AppStatus = "active"
	AppStatusDisabled AppStatus = "disabled"
	AppStatusDeleting AppStatus = "deleting"
)

var (
	AllAppStatuses = []AppStatus{AppStatusActive, AppStatusDisabled, AppStatusDeleting}
)
