package envvarserviceimpl

import "strings"

func (s *service) HasSecretRef(v string) bool {
	// TODO: do we need to handle case `${{secrets.XXX}}`
	return strings.Contains(v, "${secrets.")
}
