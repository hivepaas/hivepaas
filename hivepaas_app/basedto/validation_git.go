package basedto

import (
	"github.com/gitsight/go-vcsurl"
	vld "github.com/tiendc/go-validator"

	gitvalidation "github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper/validation"
)

func ValidateGitRepoURL[T ~string](s *T, required bool, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		_, err := vcsurl.Parse(string(*s))
		if err != nil {
			result = append(result, vld.Must(false).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_URL_INVALID"),
			))
		}
	}
	return result
}

func ValidateGitCommitHash[T ~string](s *T, required bool, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		if !gitvalidation.IsCommitHash(string(*s)) {
			result = append(result, vld.Must(false).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
			))
		}
	}
	return result
}
