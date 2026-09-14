package appdto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validReq() *GetAppLogHistoryReq {
	return &GetAppLogHistoryReq{
		ProjectID: "01J0000000000000000000000P", ProjectEnvID: "01J000000000000000000000PE",
		AppID: "01J0000000000000000000000A",
	}
}

func TestLogHistoryReqDefaults(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	req := validReq()
	req.ApplyDefaults(now)
	assert.Equal(t, now, req.End)
	assert.Equal(t, now.Add(-time.Hour), req.Start)
	assert.Equal(t, DefaultLogHistoryLimit, req.Limit)
}

func TestLogHistoryReqValidate(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		mod  func(r *GetAppLogHistoryReq)
		ok   bool
	}{
		{"defaults", func(*GetAppLogHistoryReq) {}, true},
		{"levels", func(r *GetAppLogHistoryReq) { r.Levels = []string{"error", " WARN"} }, true},
		{"unknown level", func(r *GetAppLogHistoryReq) { r.Levels = []string{"error", "verbose"} }, false},
		{"streams", func(r *GetAppLogHistoryReq) { r.Streams = []string{"stderr"} }, true},
		{"unknown stream", func(r *GetAppLogHistoryReq) { r.Streams = []string{"stdin"} }, false},
		{"limit too big", func(r *GetAppLogHistoryReq) { r.Limit = 5001 }, false},
		{"limit negative", func(r *GetAppLogHistoryReq) { r.Limit = -1 }, false},
		{"search too long", func(r *GetAppLogHistoryReq) { r.Search = string(make([]byte, 257)) }, false},
		{"end before start", func(r *GetAppLogHistoryReq) { r.Start = now; r.End = now.Add(-time.Minute) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := validReq()
			tc.mod(r)
			// The handler modifies before it validates; a case that reads as
			// valid only without that step is not the behavior being checked.
			assert.NoError(t, r.ModifyRequest())
			errs := r.Validate()
			if tc.ok {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

// Splitting is the decoder's job - see TestParseQuerySplitsListsEitherWayTheyAreWritten.
// What is left here is folding the pieces to the spelling the allowed values
// use, which ModifyRequest does before the handler validates.
func TestLogHistoryReqModifyRequestNormalizesLists(t *testing.T) {
	r := validReq()
	r.Levels = []string{" error", "WARN ", ""}
	r.Streams = []string{"stderr"}

	assert.NoError(t, r.ModifyRequest())

	assert.Equal(t, []string{"error", "warn"}, r.Levels)
	assert.Equal(t, []string{"stderr"}, r.Streams)
	assert.Empty(t, r.Validate(), "the folded values are the allowed ones")
}
