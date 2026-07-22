package envvarserviceimpl

import "strings"

func (s *service) HasRef(v string) bool {
	return strings.Contains(v, "${")
}

func (s *service) HasSecretRef(v string) bool {
	return strings.Contains(v, "${secrets.")
}
