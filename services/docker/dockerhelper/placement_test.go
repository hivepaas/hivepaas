package dockerhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlacementConstraint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expectK  string
		expectOp string
		expectV  string
	}{
		{
			name:     "valid equal constraint with spaces",
			input:    "node.role == manager",
			expectK:  "node.role",
			expectOp: "==",
			expectV:  "manager",
		},
		{
			name:     "valid equal constraint without spaces",
			input:    "node.role==manager",
			expectK:  "node.role",
			expectOp: "==",
			expectV:  "manager",
		},
		{
			name:     "valid not equal constraint with spaces",
			input:    "node.labels.env != production",
			expectK:  "node.labels.env",
			expectOp: "!=",
			expectV:  "production",
		},
		{
			name:     "valid not equal constraint without spaces",
			input:    "node.labels.env!=production",
			expectK:  "node.labels.env",
			expectOp: "!=",
			expectV:  "production",
		},
		{
			name:     "valid constraint with empty value",
			input:    "node.id == ",
			expectK:  "node.id",
			expectOp: "==",
			expectV:  "",
		},
		{
			name:     "invalid constraint missing key with ==",
			input:    "== manager",
			expectK:  "",
			expectOp: "",
			expectV:  "",
		},
		{
			name:     "invalid constraint missing key with !=",
			input:    "!= worker",
			expectK:  "",
			expectOp: "",
			expectV:  "",
		},
		{
			name:     "invalid constraint without operator",
			input:    "node.role manager",
			expectK:  "",
			expectOp: "",
			expectV:  "",
		},
		{
			name:     "empty string",
			input:    "",
			expectK:  "",
			expectOp: "",
			expectV:  "",
		},
		{
			name:     "whitespace string",
			input:    "   ",
			expectK:  "",
			expectOp: "",
			expectV:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, op, v := ParsePlacementConstraint(tt.input)
			assert.Equal(t, tt.expectK, k)
			assert.Equal(t, tt.expectOp, op)
			assert.Equal(t, tt.expectV, v)
		})
	}
}
