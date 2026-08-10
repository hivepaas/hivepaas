package base

type AppStatus string

const (
	AppStatusActive   AppStatus = "active"
	AppStatusDisabled AppStatus = "disabled"
	AppStatusDeleting AppStatus = "deleting"
	AppStatusMissing  AppStatus = "missing" // this is not used in DB
)

var (
	AllAppStatuses = []AppStatus{AppStatusActive, AppStatusDisabled, AppStatusDeleting}
)
