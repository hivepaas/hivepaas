package mcp

// Tools are every tool the server registers, in the order a client lists them.
func Tools() []Tool {
	tools := endpointTools()
	return append(tools,
		getAppLogsTool(),
		getAppDeploymentLogsTool(),
		getTaskLogsTool(),
		getAppSettingsTool(),
		getEnvLinkSuggestionsTool(),
		getProjectSettingsTool(),
		preflightInstallTool(),
		explainScheduleTool(),
		planRestartAppTool(),
		planRedeployAppTool(),
		planSetAppRunningTool(),
		planCancelDeploymentTool(),
		planInstallAppTool(),
		planUpdateAppSettingsTool(),
		planUpdateProjectSettingsTool(),
		planCreateSchedJobTool(),
		planRunSchedJobTool(),
		applyPlanTool(),
	)
}
