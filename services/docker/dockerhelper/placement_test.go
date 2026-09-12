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

func TestNodeLabelConstraint(t *testing.T) {
	cases := []struct {
		label string
		op    string
		want  string
	}{
		{"zone=eu", "==", "node.labels.zone==eu"},
		{"zone=eu", "!=", "node.labels.zone!=eu"},
		{" zone = eu ", "==", "node.labels.zone==eu"},
		// A label used as a flag: presence is written as =true, which is how
		// build-node exclusions have always been read.
		{"gpu", "==", "node.labels.gpu==true"},
		{"gpu=", "!=", "node.labels.gpu!=true"},
	}
	for _, tc := range cases {
		got, ok := NodeLabelConstraint(tc.label, tc.op)
		assert.True(t, ok, tc.label)
		assert.Equal(t, tc.want, got, tc.label)
	}
}

func TestNodeLabelConstraintRefusesWhatCannotBeOne(t *testing.T) {
	// A comma would not survive the comma-joined label HivePaaS records its own
	// constraints in; a pasted operator would produce node.labels.node.labels.x.
	for _, label := range []string{"", "  ", "=eu", "zone==eu", "zone!=eu", "a=1,b=2", ",", "zone=eu,"} {
		_, ok := NodeLabelConstraint(label, "==")
		assert.False(t, ok, "label %q", label)
	}
}
