package volumeagentdto

type RemoveVolumeReq struct {
	VolumeID string
	Force    bool
}

type RemoveVolumeResp struct{}
