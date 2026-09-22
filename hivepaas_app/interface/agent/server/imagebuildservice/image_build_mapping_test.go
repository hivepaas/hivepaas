package imagebuildservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// A build that runs on another node has to tag the image the same way, so the
// deployment's tags have to survive the wire.
func TestImageTagsSurviveTheProto(t *testing.T) {
	req := &agentproto.ImageBuildReq{
		TaskId: "t1", AppId: "a1", CommitHash: "9f3c1de0ab",
		ImageTags: []string{"v1.4.0", "stable"},
	}

	assert.Equal(t, []string{"v1.4.0", "stable"}, req.GetImageTags())
}
