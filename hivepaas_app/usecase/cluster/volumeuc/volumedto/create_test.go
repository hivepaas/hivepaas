package volumedto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func volumeReq(nodeID, nodeLabel string) *VolumeBaseReq {
	return &VolumeBaseReq{
		Name:      "data",
		NodeID:    nodeID,
		NodeLabel: nodeLabel,
	}
}

func validationErrors(req *VolumeBaseReq) hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(req.validate("")...))
}

func hasFieldError(errs hperrors.ValidationErrors, field string) bool {
	for _, err := range errs {
		if err != nil && strings.Contains(err.Error(), field) {
			return true
		}
	}
	return false
}

func TestVolumeNodePinningIsOneOrTheOther(t *testing.T) {
	t.Run("by id", func(t *testing.T) {
		assert.Empty(t, validationErrors(volumeReq("node-1", "")))
	})

	t.Run("by label", func(t *testing.T) {
		assert.Empty(t, validationErrors(volumeReq("", "storage=fast")))
	})

	// Not an omission: it is the claim that the volume is reachable everywhere.
	t.Run("neither, which is a claim of its own", func(t *testing.T) {
		assert.Empty(t, validationErrors(volumeReq("", "")))
	})

	// Both would have the label quietly dropped downstream, so it is refused -
	// and both fields are named, because either one is the one to take out.
	t.Run("both is refused", func(t *testing.T) {
		errs := validationErrors(volumeReq("node-1", "storage=fast"))

		if len(errs) == 0 {
			t.Fatal("a request pinning by id and by label at once must be refused")
		}
		if !hasFieldError(errs, "nodeId") {
			t.Error("the refusal has to name nodeId")
		}
		if !hasFieldError(errs, "nodeLabel") {
			t.Error("the refusal has to name nodeLabel")
		}
	})
}

func strPtr(v string) *string { return &v }

func updateReq(nodeID, nodeLabel *string) *UpdateVolumeReq {
	req := NewUpdateVolumeReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WA9"
	req.NodeID = nodeID
	req.NodeLabel = nodeLabel
	return req
}

// Saying nothing about the pinning is what leaves it alone; the pair only moves
// when the request mentions it.
func TestUpdateVolumePinningIsOptional(t *testing.T) {
	assert.Nil(t, updateReq(nil, nil).Pinning())
	assert.Empty(t, updateReq(nil, nil).Validate())
}

// Either half mentioned replaces both, so the volume never ends up claiming to
// be pinned two ways at once.
func TestUpdateVolumePinningReplacesThePair(t *testing.T) {
	t.Run("id alone clears the label", func(t *testing.T) {
		pinning := updateReq(strPtr("node-1"), nil).Pinning()

		assert.NotNil(t, pinning)
		assert.Equal(t, "node-1", pinning.NodeID)
		assert.Empty(t, pinning.NodeLabel)
	})

	t.Run("label alone clears the id", func(t *testing.T) {
		pinning := updateReq(nil, strPtr("storage=fast")).Pinning()

		assert.NotNil(t, pinning)
		assert.Empty(t, pinning.NodeID)
		assert.Equal(t, "storage=fast", pinning.NodeLabel)
	})

	// Both empty is an answer, not an omission: reachable from every node.
	t.Run("both empty unpins the volume", func(t *testing.T) {
		req := updateReq(strPtr(""), strPtr(""))
		pinning := req.Pinning()

		assert.NotNil(t, pinning)
		assert.Empty(t, pinning.NodeID)
		assert.Empty(t, pinning.NodeLabel)
		assert.Empty(t, req.Validate())
	})
}

func TestUpdateVolumeRefusesBothWaysAtOnce(t *testing.T) {
	errs := updateReq(strPtr("node-1"), strPtr("storage=fast")).Validate()

	if len(errs) == 0 {
		t.Fatal("pinning by id and by label at once must be refused on update too")
	}
	assert.True(t, hasFieldError(errs, "nodeId"))
	assert.True(t, hasFieldError(errs, "nodeLabel"))
}
