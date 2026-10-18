package artifactagentdto

// EnsureReq is a file to have on the node: where it is published, an optional
// mirror tried first, and what it must hash to and weigh.
type EnsureReq struct {
	URL       string
	MirrorURL string
	SHA256    string
	Size      int64
}

// EnsureResp is where the file is on the host, which is what a bind mount names,
// and whether this call downloaded it.
type EnsureResp struct {
	HostPath   string
	Downloaded bool
}
