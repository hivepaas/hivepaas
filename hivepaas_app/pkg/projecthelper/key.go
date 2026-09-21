package projecthelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
)

const (
	objectKeyMaxLen = 100

	// AppKeyMaxLen is the longest a DNS label may be. An app's key is its name on
	// the project network - the container hostname, the network alias and the
	// value of HIVEPAAS_HOST - so it has to be one.
	AppKeyMaxLen = 63
)

func CalcProjectKey(projectName string) string {
	return slugify.SlugifyEx(projectName, []string{"-", "_"}, objectKeyMaxLen)
}

// CalcAppKey derives an app's key from its name: the slug, with the hyphens the
// slug is made of kept as they are.
//
// The key is also the app's host name on the project network, and a host name
// may hold letters, digits and hyphens only. Underscores, which keys used to be
// written with, resolve through docker's DNS but are refused by every strict
// parser an app brings: Django and Tomcat answer 400 to such a Host, image
// references reject one in a registry name, Rails cannot parse a database URL
// pointing at it, and the AWS SDKs refuse it as an endpoint.
func CalcAppKey(appName string) string {
	key := slugify.SlugifyEx(appName, nil, AppKeyMaxLen)
	// Cutting to length can leave the hyphen that separated two words at the end,
	// which a DNS label may not end with.
	return strings.TrimRight(key, "-")
}

func CalcAppGlobalKey(projectKey, appKey, env string) string {
	globalKey := projectKey
	if env != "" {
		globalKey += "_" + CalcProjectEnvKey(env)
	}
	globalKey += "_" + appKey
	return globalKey
}
