package templatemodel

import (
	"encoding/json"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// TrendingWindowDays is the span StarsGained counts over.
const TrendingWindowDays = 28

// DateLayout is how the stats write a day.
const DateLayout = time.DateOnly

// IndexStats are what the store orders templates by besides their name: when a
// template joined the repository, and how much its upstream project is starred.
//
// They come from stats.yaml, not from the template file, so that changing them
// never touches a file an installation reads strictly. And they are read
// leniently: a figure this HivePaaS cannot read is a template sorted as one
// without it, never a template left out of the store.
type IndexStats struct {
	// Added is the day the template joined the repository, YYYY-MM-DD.
	Added string `json:"added,omitempty"`
	// Stars are the GitHub stars of the project the template runs, when they
	// were last counted. Zero is a project not counted - not on GitHub, or not
	// found there.
	Stars int `json:"stars,omitempty"`
	// StarsGained are the stars it gained in the TrendingWindowDays before.
	StarsGained int `json:"starsGained,omitempty"`
}

// UnmarshalJSON reads each figure on its own and does without one it cannot
// read; it never fails.
func (s *IndexStats) UnmarshalJSON(data []byte) error {
	*s = IndexStats{}
	// Stats that are not an object leave fields empty, and every figure unknown.
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(data, &fields)
	var added string
	if json.Unmarshal(fields["added"], &added) == nil && ValidDate(added) {
		s.Added = added
	}
	s.Stars = count(fields["stars"])
	s.StarsGained = count(fields["starsGained"])
	return nil
}

func count(raw json.RawMessage) int {
	var n int
	if json.Unmarshal(raw, &n) != nil || n < 0 {
		return 0
	}
	return n
}

// ValidDate reports whether value is a day written as DateLayout.
func ValidDate(value string) bool {
	_, err := time.Parse(DateLayout, value)
	return err == nil
}

// Stats is stats.yaml, which the tooling writes and the index takes IndexStats
// from: `apptemplate index` dates a template the day it joins, and `apptemplate
// stats` counts the stars every week. Installations never read it.
type Stats struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	// Counted is the day the stars were last counted.
	Counted   string                    `yaml:"counted,omitempty"`
	Templates map[string]*TemplateStats `yaml:"templates"`
}

type TemplateStats struct {
	Added string `yaml:"added"`
	// Repo is the GitHub repository whose stars count, as owner/name, for a
	// template whose source link is not one - a project on Codeberg with a
	// mirror on GitHub. It is the only field written by hand.
	Repo        string `yaml:"repo,omitempty"`
	Stars       int    `yaml:"stars,omitempty"`
	StarsGained int    `yaml:"starsGained,omitempty"`
}

func DecodeStats(data []byte) (*Stats, error) {
	stats := &Stats{}
	if err := decodeYAMLStrict(data, stats); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("stats.yaml: %s", err.Error())
	}
	if err := checkHeader("stats.yaml", stats.APIVersion, stats.Kind, KindTemplateStats); err != nil {
		return nil, err
	}
	if stats.Templates == nil {
		stats.Templates = map[string]*TemplateStats{}
	}
	return stats, nil
}
