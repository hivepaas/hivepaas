package entity

// OBIAgentSettings are what an agent runs OBI by, read every 30 seconds on
// every node: whether the logs are stored, the feature's switch and nodes, and
// the apps that ask for their routes and calls, by their swarm service's name.
// They are cached in Redis for the agents, and read from the database on a
// miss and every 10 minutes.
type OBIAgentSettings struct {
	LogsOn      bool                `json:"logsOn"`
	Performance *LoggingPerformance `json:"performance,omitempty"`
	// Apps are read only while the feature is on: none while it is off.
	Apps map[string]string `json:"apps,omitempty"`
}

// On says the feature runs: on, and the logs stored.
func (s *OBIAgentSettings) On() bool {
	return s != nil && s.LogsOn && s.Performance != nil && s.Performance.Enabled
}
