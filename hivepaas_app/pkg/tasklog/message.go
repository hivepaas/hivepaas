package tasklog

import (
	"strings"
)

type Command string

const (
	CommandNewData Command = "new-data"
	CommandClosed  Command = "closed"
)

func parseMessage(msg string) (Command, any) {
	cmd, data, _ := strings.Cut(msg, "\n")
	return Command(cmd), data
}

func buildMessage(cmd Command) any {
	return string(cmd)
}
