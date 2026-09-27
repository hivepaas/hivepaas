package config

import "time"

type Agent struct {
	Port        int    `toml:"port" env:"HP_AGENT_PORT" default:"10001"`
	SecretToken string `toml:"secret_token" env:"HP_AGENT_SECRET_TOKEN"`
	// RepoServer is the kopia repository server the agent runs for the time of a
	// backup into a repository on one of its node's volumes.
	RepoServer AgentRepoServer `toml:"repo_server"`
}

type AgentRepoServer struct {
	// MemLimit is the server's GOMEMLIMIT: without one, it keeps near 1 GB while
	// it takes a stream.
	MemLimit string `toml:"mem_limit" env:"HP_AGENT_REPO_SERVER_MEM_LIMIT" default:"256MiB"`
	// StartTimeout is how long the server has to listen before it is stopped.
	StartTimeout time.Duration `toml:"start_timeout" env:"HP_AGENT_REPO_SERVER_START_TIMEOUT" default:"30s"`
}
