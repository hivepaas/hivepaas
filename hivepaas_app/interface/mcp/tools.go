package mcp

// Tools are every tool the server registers, in the order a client lists them.
func Tools() []Tool {
	return []Tool{
		listProjectsTool(),
		listAppsTool(),
		getAppTool(),
		getAppStatusTool(),
		getAppLogsTool(),
		listAppDeploymentsTool(),
		getAppDeploymentTool(),
		getAppSettingsTool(),
		listAttentionTool(),
		listTasksTool(),
		getTaskLogsTool(),
		listNodesTool(),
		searchTemplatesTool(),
		getTemplateTool(),
		preflightInstallTool(),
		listSchedJobsTool(),
		explainScheduleTool(),
		planRestartAppTool(),
		planRedeployAppTool(),
		planSetAppRunningTool(),
		planCancelDeploymentTool(),
		planInstallAppTool(),
		planUpdateAppSettingsTool(),
		planCreateSchedJobTool(),
		applyPlanTool(),
	}
}
