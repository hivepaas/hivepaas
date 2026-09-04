package userdto

import (
	"testing"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

func envAccessReq(id string) *basedto.ObjectAccessReq {
	return &basedto.ObjectAccessReq{
		ObjectIDReq: basedto.ObjectIDReq{ID: id},
		Access:      base.AccessActions{Read: true},
	}
}

func TestValidateEnvAccesses(t *testing.T) {
	tests := []struct {
		name     string
		accesses basedto.ObjectAccessSliceReq
		wantErr  bool
	}{
		{
			name:     "empty list is allowed",
			accesses: nil,
		},
		{
			name: "project env IDs are accepted",
			accesses: basedto.ObjectAccessSliceReq{
				envAccessReq("prj_1:dev"), envAccessReq("prj_1:prod"),
			},
		},
		{
			// Permissions are per env now: a bare project ID would silently grant
			// access to every env of the project.
			name:     "a bare project ID is rejected",
			accesses: basedto.ObjectAccessSliceReq{envAccessReq("prj_1")},
			wantErr:  true,
		},
		{
			name: "one bad ID among good ones is still rejected",
			accesses: basedto.ObjectAccessSliceReq{
				envAccessReq("prj_1:dev"), envAccessReq("prj_2"),
			},
			wantErr: true,
		},
		{
			name:     "empty ID is rejected",
			accesses: basedto.ObjectAccessSliceReq{envAccessReq("")},
			wantErr:  true,
		},
		{
			name: "duplicated envs are rejected",
			accesses: basedto.ObjectAccessSliceReq{
				envAccessReq("prj_1:dev"), envAccessReq("prj_1:dev"),
			},
			wantErr: true,
		},
		{
			// The grant is stored from the env ID alone, so an env belonging to
			// another project would end up granted under the wrong project.
			name:     "an env of another project is rejected",
			accesses: basedto.ObjectAccessSliceReq{envAccessReq("prj_2:dev")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vld.Validate(validateEnvAccesses(tt.accesses, "prj_1", "envAccesses")...)
			if tt.wantErr && len(err) == 0 {
				t.Error("expected a validation error, got none")
			}
			if !tt.wantErr && len(err) > 0 {
				t.Errorf("expected no validation error, got %v", err)
			}
		})
	}
}
