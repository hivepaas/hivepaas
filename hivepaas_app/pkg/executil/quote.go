package executil

import (
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func IsSingleQuoted(s string) bool {
	return strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")
}

func IsDoubleQuoted(s string) bool {
	return strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

func IsQuoted(s string) bool {
	return IsSingleQuoted(s) || IsDoubleQuoted(s)
}

func ArgQuote(arg string) string {
	if IsQuoted(arg) {
		return arg
	}
	// Do not escape template placeholders ${...} or single words without whitespace/quotes
	if strings.Contains(arg, "${") || !strings.ContainsAny(arg, " \t\n\"'") {
		return arg
	}
	return shellquote.Join(arg)
}

func CmdSplit(cmd string) ([]string, error) {
	res, err := shellquote.Split(cmd)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return res, nil
}
