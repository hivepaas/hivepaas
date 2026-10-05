package appmetricsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAMetricsRangeIsOneOfFour(t *testing.T) {
	for _, r := range []string{"1h", "6h", "24h", "7d"} {
		req := &GetFunctionMetricsReq{ProjectID: "01JAB9XED0GTXBSQDFVYAJ8WB1", ProjectEnvID: "dev",
			AppID: "01JAB9XED0GTXBSQDFVYAJ8WD1", Range: r}
		assert.Empty(t, req.Validate(), r)
	}
	req := &GetFunctionMetricsReq{ProjectID: "01JAB9XED0GTXBSQDFVYAJ8WB1", ProjectEnvID: "dev",
		AppID: "01JAB9XED0GTXBSQDFVYAJ8WD1", Range: "2h"}
	assert.NotEmpty(t, req.Validate())

	req = NewGetFunctionMetricsReq()
	assert.Equal(t, DefaultFunctionMetricsRange, req.Range, "a day, when the range is not given")
	assert.Equal(t, "24h", DefaultFunctionMetricsRange)
}
