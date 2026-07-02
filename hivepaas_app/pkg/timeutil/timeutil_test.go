package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNowUTC(t *testing.T) {
	now := NowUTC()
	assert.Equal(t, time.UTC, now.Location())
	assert.True(t, time.Now().After(now.Add(-time.Second)))
}

func TestCurrentDateUTC(t *testing.T) {
	date := CurrentDateUTC()
	assert.Equal(t, time.UTC, date.ToTime().Location())
	now := time.Now().UTC()
	assert.Equal(t, now.Year(), date.ToTime().Year())
	assert.Equal(t, now.Month(), date.ToTime().Month())
	assert.Equal(t, now.Day(), date.ToTime().Day())
	assert.Equal(t, 0, date.ToTime().Hour())
}

func TestCurrentYearUTC(t *testing.T) {
	year := CurrentYearUTC()
	assert.Equal(t, time.Now().UTC().Year(), year)
}
