package mcp

// Tools are every tool the server registers, in the order a client lists them.
func Tools() []Tool {
	return []Tool{
		listProjectsTool(),
		listAppsTool(),
		getAppTool(),
		getAppStatusTool(),
		getAppLogsTool(),
		getAppConfigTool(),
	}
}
