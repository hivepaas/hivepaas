package volumedto

import (
	"reflect"
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

func updateReq() *UpdateVolumeReq {
	req := NewUpdateVolumeReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WA9"
	return req
}

// Pinning is settled when the volume is created. An update that could move it
// would be a promise HivePaaS cannot keep: the data does not follow the pin, so
// the only thing a change moves is where HivePaaS goes looking.
func TestUpdateVolumeCarriesNoPinning(t *testing.T) {
	assert.Empty(t, updateReq().Validate())

	req := updateReq()
	typ := reflect.TypeOf(*req)
	for _, field := range []string{"NodeID", "NodeLabel"} {
		if _, found := typ.FieldByName(field); found {
			t.Errorf("update request must not carry %s: pinning is immutable", field)
		}
	}
}
