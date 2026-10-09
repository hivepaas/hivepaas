package tasklog

import (
	"strings"
)

type Command string

const (
	CommandNewData Command = "new-data"
	CommandClosed  Command = "closed"
	// CommandReset is a log whose list started again, its frames handed on:
	// those following read on from the start of the new list.
	CommandReset Command = "reset"
)

func parseMessage(msg string) (Command, any) {
	cmd, data, _ := strings.Cut(msg, "\n")
	return Command(cmd), data
}

func buildMessage(cmd Command) any {
	return string(cmd)
}
