package hperrors

// Errors for artifacts: the files a managed template's engine downloads to a node
// and mounts into an app's container.
var (
	// ErrArtifactInvalid refuses an artifact that cannot be trusted as asked for:
	// not over https, a malformed hash, a size out of bounds.
	ErrArtifactInvalid = NewErr(ErrValueInvalid, "ERR_ARTIFACT_INVALID")
	// ErrArtifactHashMismatch is a download that is not the file its hash names.
	// Nothing of it is kept.
	ErrArtifactHashMismatch = NewErr(ErrValueInvalid, "ERR_ARTIFACT_HASH_MISMATCH")
	// ErrArtifactDownloadFailed is a download that did not complete: the source
	// could not be reached, answered with an error, sent too much or too little.
	ErrArtifactDownloadFailed = NewErr(ErrUnavailable, "ERR_ARTIFACT_DOWNLOAD_FAILED")
	// ErrArtifactArchUnsupported is a node whose architecture the artifact has no
	// file for.
	ErrArtifactArchUnsupported = NewErr(ErrPreconditionFailed, "ERR_ARTIFACT_ARCH_UNSUPPORTED")
)
