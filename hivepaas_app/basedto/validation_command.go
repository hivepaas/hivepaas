package basedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
)

// ValidateCommandLine checks that a command line splits by shell rules: one
// with a quote left open does not, and would fail every deployment.
func ValidateCommandLine(line *string, field string) []vld.Validator {
	if line == nil || *line == "" {
		return nil
	}
	_, err := executil.CmdSplit(*line)
	return ValidateCond(err == nil, field)
}
