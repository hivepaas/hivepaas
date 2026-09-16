package specserviceimpl

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// yamlIndent is two spaces. yaml.v3 defaults to four, which reads badly at the
// nesting depth an app reaches.
const yamlIndent = 2

// marshalDoc writes one payload file.
//
// yaml.v3 sorts map keys already, which is what makes two exports of unchanged
// data byte-identical: Go's randomized map iteration would otherwise reorder
// every label block between runs and make every diff noise. Struct fields come
// out in declaration order, so ordering within a document is decided by how the
// types in specmodel are written.
func marshalDoc(doc any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndent)

	if err := encoder.Encode(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := encoder.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
