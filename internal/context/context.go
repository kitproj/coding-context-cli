package context

import "os"

type Config struct {
	AgentToken string
	Endpoint   string
}

func LoadFromEnv() Config {
	return Config{
		AgentToken: os.Getenv("AGENT_TOKEN"),
		Endpoint:   os.Getenv("AGENT_ENDPOINT"),
	}
}
