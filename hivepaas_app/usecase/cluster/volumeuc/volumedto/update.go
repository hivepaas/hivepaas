package volumedto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateVolumeReq struct {
	settings.UpdateSettingReq

	// Where the volume's data is - see VolumeBaseReq.NodeID for what the pair
	// means and what it decides.
	//
	// Pointers, and the two move together: leaving both out keeps the pinning the
	// volume already has, while sending either one replaces the pair, reading the
	// one left out as empty. They are two spellings of a single answer, so
	// updating one half on its own would leave a volume claiming to be pinned two
	// ways at once.
	//
	// Sending both empty is an answer too, not a way of skipping the field: it
	// says the volume is reachable from every node.
	NodeID    *string `json:"nodeId"`
	NodeLabel *string `json:"nodeLabel"`
}

func NewUpdateVolumeReq() *UpdateVolumeReq {
	return &UpdateVolumeReq{}
}

// Pinning is the node pinning this request asks for, or nil when it says nothing
// about it.
func (req *UpdateVolumeReq) Pinning() *entity.ClusterVolume {
	if req.NodeID == nil && req.NodeLabel == nil {
		return nil
	}
	return &entity.ClusterVolume{
		NodeID:    gofn.PtrValueOrEmpty(req.NodeID),
		NodeLabel: gofn.PtrValueOrEmpty(req.NodeLabel),
	}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateVolumeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)

	// The same rule as on create: one way of naming the node or the other, never
	// both, because the node exec service takes the id and the label would be
	// dropped without a word.
	if pinning := req.Pinning(); pinning != nil {
		validators = append(validators, basedto.ValidateMutualExclusiveFields(
			pinning.NodeID == "" || pinning.NodeLabel == "",
			"nodeId", "nodeLabel")...)
	}

	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
