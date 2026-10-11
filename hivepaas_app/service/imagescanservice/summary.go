package imagescanservice

import (
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Ignore is a vulnerability ignore, with the setting it is: its scope, who
// wrote it and when.
type Ignore struct {
	Setting *entity.Setting
	Data    *entity.VulnerabilityIgnore
}

// IgnoresOf picks the ignores that apply to an app out of settings: its own,
// its parent app's, and those of its env, its project and the installation
// that are inherited. A disabled one, or one that cannot be read, applies to
// nothing.
func IgnoresOf(app *entity.App, settings []*entity.Setting) []*Ignore {
	var out []*Ignore
	for _, setting := range settings {
		if setting.Type != base.SettingTypeVulnerabilityIgnore || setting.Status != base.SettingStatusActive {
			continue
		}
		if !appliesTo(setting, app) {
			continue
		}
		data, err := setting.AsVulnerabilityIgnore()
		if err != nil {
			continue
		}
		out = append(out, &Ignore{Setting: setting, Data: data})
	}
	return out
}

func appliesTo(setting *entity.Setting, app *entity.App) bool {
	switch setting.Scope {
	case base.ObjectScopeApp:
		return setting.ObjectID == app.ID || (setting.Inheritable && app.ParentID != "" &&
			setting.ObjectID == app.ParentID)
	case base.ObjectScopeProjectEnv:
		return setting.Inheritable && setting.ObjectID == app.ProjectEnvID
	case base.ObjectScopeProject:
		return setting.Inheritable && setting.ObjectID == app.ProjectID
	case base.ObjectScopeGlobal:
		return setting.Inheritable
	case base.ObjectScopeHivepaas, base.ObjectScopeUser:
		return false
	}
	return false
}

// Applied is the ignore that sets the finding aside at a time: the nearest
// scope's when several do. Nil when none does. One past its setting's
// ExpireAt applies no more.
func Applied(f *entity.ImageScanFinding, ignores []*Ignore, at time.Time) *Ignore {
	var applied *Ignore
	for _, ignore := range ignores {
		if !ignore.Data.Applies(f.VulnID, f.PkgName) {
			continue
		}
		if expireAt := ignore.Setting.ExpireAt; !expireAt.IsZero() && !at.Before(expireAt) {
			continue
		}
		if applied == nil || scopeRank(ignore.Setting.Scope) < scopeRank(applied.Setting.Scope) {
			applied = ignore
		}
	}
	return applied
}

// scopeRank orders scopes from the nearest to an app to the farthest.
func scopeRank(scope base.ObjectScopeType) int {
	ranks := []base.ObjectScopeType{base.ObjectScopeApp, base.ObjectScopeProjectEnv, base.ObjectScopeProject,
		base.ObjectScopeGlobal}
	if i := slices.Index(ranks, scope); i >= 0 {
		return i
	}
	return len(ranks)
}

// Count is how many findings of a severity, and how many of them a fixed
// version exists for.
type Count struct {
	Total   int `json:"total"`
	Fixable int `json:"fixable"`
}

// Summary counts findings by severity, after ignores.
type Summary struct {
	Critical Count `json:"critical"`
	High     Count `json:"high"`
	Medium   Count `json:"medium"`
	Low      Count `json:"low"`
	Unknown  Count `json:"unknown"`
	// Ignored is how many were set aside, of every severity.
	Ignored int `json:"ignored"`
}

func (s *Summary) countOf(severity base.VulnSeverity) *Count {
	switch severity {
	case base.VulnSeverityCritical:
		return &s.Critical
	case base.VulnSeverityHigh:
		return &s.High
	case base.VulnSeverityMedium:
		return &s.Medium
	case base.VulnSeverityLow:
		return &s.Low
	case base.VulnSeverityUnknown:
		return &s.Unknown
	}
	return &s.Unknown
}

// Add counts one more finding in.
func (s *Summary) Add(f *entity.ImageScanFinding) {
	count := s.countOf(f.Severity)
	count.Total++
	if f.Fixable() {
		count.Fixable++
	}
}

// Merge counts another summary in.
func (s *Summary) Merge(other Summary) {
	for _, severity := range base.AllVulnSeverities {
		mine, theirs := s.countOf(severity), other.countOf(severity)
		mine.Total += theirs.Total
		mine.Fixable += theirs.Fixable
	}
	s.Ignored += other.Ignored
}

// Summarize counts findings that no ignore sets aside at a time, and those
// that one does.
func Summarize(findings []*entity.ImageScanFinding, ignores []*Ignore, at time.Time) Summary {
	var summary Summary
	for _, f := range findings {
		if Applied(f, ignores, at) != nil {
			summary.Ignored++
			continue
		}
		summary.Add(f)
	}
	return summary
}

// Reportable are the keys of the findings the daily scan reports when they are
// new: fixable CRITICAL and HIGH, which no ignore sets aside at a time. Sorted.
func Reportable(findings []*entity.ImageScanFinding, ignores []*Ignore, at time.Time) []string {
	var keys []string
	for _, f := range findings {
		if !f.Fixable() || f.Severity.Rank() > base.VulnSeverityHigh.Rank() {
			continue
		}
		if Applied(f, ignores, at) != nil {
			continue
		}
		keys = append(keys, f.Key())
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}
