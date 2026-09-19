package dockerhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageProvidesInit(t *testing.T) {
	cases := map[string]struct {
		entrypoint []string
		want       bool
	}{
		"nothing at all":             {nil, false},
		"an empty entry point":       {[]string{}, false},
		"tini by name":               {[]string{"tini", "--"}, true},
		"tini by path":               {[]string{"/usr/bin/tini", "-g", "--"}, true},
		"tini in front of a shell":   {[]string{"tini", "--", "/bin/bash", "-c"}, true},
		"dumb-init":                  {[]string{"/usr/bin/dumb-init", "--"}, true},
		"catatonit":                  {[]string{"/usr/bin/catatonit", "--"}, true},
		"s6-overlay":                 {[]string{"/init"}, true},
		"an ordinary entry point":    {[]string{"docker-entrypoint.sh"}, false},
		"a binary of the app's own":  {[]string{"/app/server", "--serve"}, false},
		"something that only sounds": {[]string{"/usr/bin/initialize-db"}, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, ImageProvidesInit(tc.entrypoint))
		})
	}
}
