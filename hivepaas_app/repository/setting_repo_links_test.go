package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func settingLink(dstID string) *entity.ResLink {
	return &entity.ResLink{
		SrcType: base.ResourceTypeSetting, SrcID: "JOB",
		DstType: base.ResourceTypeSetting, DstID: dstID,
	}
}

func linkKeys(links []*entity.ResLink) []string {
	keys := make([]string, 0, len(links))
	for _, link := range links {
		keys = append(keys, link.GetKey())
	}
	return keys
}

// A target told of success and of failure is wanted twice. One statement may
// not upsert a row twice: Postgres refuses it (21000), and the whole save with
// it - a health check whose two notifications go to the same Slack answered 500.
func TestResLinksToUpsertWritesALinkWantedTwiceOnce(t *testing.T) {
	now := time.Now()

	links := resLinksToUpsert([]*entity.ResLink{settingLink("SLACK"), settingLink("SLACK")}, nil, now)
	assert.Equal(t, []string{settingLink("SLACK").GetKey()}, linkKeys(links))

	// The same when the link is there already, and changed.
	current := settingLink("SLACK")
	current.Index = 1
	links = resLinksToUpsert([]*entity.ResLink{settingLink("SLACK"), settingLink("SLACK")},
		[]*entity.ResLink{current}, now)
	assert.Equal(t, []string{settingLink("SLACK").GetKey()}, linkKeys(links))
}

func TestResLinksToUpsert(t *testing.T) {
	now := time.Now()
	kept := settingLink("KEPT")
	gone := settingLink("GONE")

	links := resLinksToUpsert([]*entity.ResLink{settingLink("KEPT"), settingLink("NEW")},
		[]*entity.ResLink{kept, gone}, now)

	// Unchanged, a link is not written again; new, it is; no longer wanted, it
	// is deleted.
	assert.ElementsMatch(t, []string{settingLink("NEW").GetKey(), gone.GetKey()}, linkKeys(links))
	assert.True(t, kept.DeletedAt.IsZero())
	assert.Equal(t, now, gone.DeletedAt)
}
