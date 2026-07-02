package executil

import (
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
	return shellquote.Join(arg)
}

func CmdSplit(cmd string) ([]string, error) {
	res, err := shellquote.Split(cmd)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return res, nil
}
