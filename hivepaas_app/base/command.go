package base

type CommandTemplateKind string

const (
	CommandTemplateBackup      CommandTemplateKind = "backup"
	CommandTemplateDataOps     CommandTemplateKind = "data-ops"
	CommandTemplateDatabase    CommandTemplateKind = "database"
	CommandTemplateDeployment  CommandTemplateKind = "deployment"
	CommandTemplateDiagnostics CommandTemplateKind = "diagnostics"
	CommandTemplateMaintenance CommandTemplateKind = "maintenance"
	CommandTemplateTesting     CommandTemplateKind = "testing"
)

var (
	AllCommandTemplateKinds = []CommandTemplateKind{CommandTemplateBackup, CommandTemplateDataOps,
		CommandTemplateDatabase, CommandTemplateDeployment, CommandTemplateDiagnostics,
		CommandTemplateMaintenance, CommandTemplateTesting}
)
