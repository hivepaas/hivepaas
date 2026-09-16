package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// Two runs over unchanged data must produce identical bytes, or the format is
// useless for review and for git.
func TestMarshalDocIsByteIdenticalAcrossRuns(t *testing.T) {
	doc := &specmodel.EnvDoc{
		DocHeader: specmodel.NewDocHeader("project-env"),
		Project:   "project_a",
		Env:       "dev",
	}

	first, err := marshalDoc(doc)
	assert.NoError(t, err)
	second, err := marshalDoc(doc)
	assert.NoError(t, err)
	assert.Equal(t, string(first), string(second))
}

// Go map iteration is randomized, so a map written straight out would differ
// between runs. Twenty runs make an unsorted encoder fail with overwhelming
// probability.
func TestMarshalDocSortsMapKeys(t *testing.T) {
	doc := &specmodel.EnvDoc{
		DocHeader: specmodel.NewDocHeader("project-env"),
		Labels: map[string]string{
			"zebra": "1", "alpha": "2", "mu": "3", "beta": "4", "omega": "5",
			"kappa": "6", "delta": "7", "sigma": "8",
		},
	}

	want, err := marshalDoc(doc)
	assert.NoError(t, err)
	for range 20 {
		got, err := marshalDoc(doc)
		assert.NoError(t, err)
		assert.Equal(t, string(want), string(got))
	}
}

func TestMarshalDocUsesTwoSpaceIndent(t *testing.T) {
	doc := &specmodel.EnvDoc{
		DocHeader: specmodel.NewDocHeader("project-env"),
		Labels:    map[string]string{"a": "1"},
	}
	out, err := marshalDoc(doc)
	assert.NoError(t, err)
	assert.Contains(t, string(out), "\n  a: \"1\"")
}

// The header is inlined rather than nested, so a file opens with what it is.
func TestMarshalDocInlinesTheHeader(t *testing.T) {
	out, err := marshalDoc(&specmodel.GlobalDoc{DocHeader: specmodel.NewDocHeader("global")})
	assert.NoError(t, err)
	assert.Equal(t, "apiVersion: hivepaas.com/v1\nkind: Spec\nscope: global\n", string(out))
}
