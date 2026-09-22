package base

const (
	// ImageNameMaxLen is what GetAutoImageName still cuts an app key to. An
	// image's name is a function of its app now - entity.App.ImageRepoName - and
	// both this and that function go in the change that takes its last caller
	// away.
	ImageNameMaxLen = 200

	// ImageRepoNameMaxLen is the OCI limit for a whole repository name, which
	// <address>/<username>/<name> has to fit inside.
	ImageRepoNameMaxLen = 255
	// ImageTagMaxLen is the OCI limit for a tag.
	ImageTagMaxLen = 128
	// ImageCustomTagMaxLen is what a deployment's own tag may hold before the
	// environment prefix is added to it.
	ImageCustomTagMaxLen = 100
	// ImageMaxCustomTags is how many a deployment may add. A release marker or
	// two is the case this exists for.
	ImageMaxCustomTags = 5
)
