package base

const (
	// ImageNameMaxLen bounds the imageName field of the deployment settings,
	// which the change that takes naming out of those settings removes.
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
