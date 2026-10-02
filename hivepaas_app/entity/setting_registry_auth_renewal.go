package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentRegistryAuthRenewalVersion = 1

	// RegistryAuthRenewalIntervalDefault is how often the ECR tokens Swarm keeps
	// are renewed when nobody says otherwise; Min and Max bound what may be said.
	// An ECR token lives 12 hours, and one handed over must outlive the next run
	// by an hour: past 10 hours no token is fresh enough.
	RegistryAuthRenewalIntervalDefault = 6 * time.Hour
	RegistryAuthRenewalIntervalMin     = time.Hour
	RegistryAuthRenewalIntervalMax     = 10 * time.Hour
)

var _ = registerSettingParser(base.SettingTypeRegistryAuthRenewal, &registryAuthRenewalParser{})

type registryAuthRenewalParser struct {
}

func (s *registryAuthRenewalParser) New() SettingData {
	return &RegistryAuthRenewal{}
}

// RegistryAuthRenewal is when the Amazon ECR tokens Swarm keeps in services are
// renewed - an interval, never a cron - and who is told how a run went.
type RegistryAuthRenewal struct {
	Schedule     SchedJobSchedule       `json:"schedule"`
	Notification *BaseEventNotification `json:"notification,omitempty"`
}

// Interval is the schedule's, kept within what a 12-hour token allows; the
// default when there is none.
func (s *RegistryAuthRenewal) Interval() time.Duration {
	interval := s.Schedule.Interval.ToDuration()
	switch {
	case interval <= 0:
		return RegistryAuthRenewalIntervalDefault
	case interval < RegistryAuthRenewalIntervalMin:
		return RegistryAuthRenewalIntervalMin
	case interval > RegistryAuthRenewalIntervalMax:
		return RegistryAuthRenewalIntervalMax
	}
	return interval
}

// NewRegistryAuthRenewal is the renewal as a new installation has it: every 6
// hours from the hour after now, failures told to the default notification.
func NewRegistryAuthRenewal(timeNow time.Time) *RegistryAuthRenewal {
	return &RegistryAuthRenewal{
		Schedule: SchedJobSchedule{
			Interval:    timeutil.Duration(RegistryAuthRenewalIntervalDefault),
			InitialTime: timeNow.Truncate(time.Hour).Add(time.Hour),
		},
		Notification: &BaseEventNotification{FailureUseDefault: true},
	}
}

func (s *RegistryAuthRenewal) GetType() base.SettingType {
	return base.SettingTypeRegistryAuthRenewal
}

func (s *RegistryAuthRenewal) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *RegistryAuthRenewal) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *RegistryAuthRenewal) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRegistryAuthRenewalVersion {
		return false, nil
	}
	if setting.Version > CurrentRegistryAuthRenewalVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}
	setting.Version = CurrentRegistryAuthRenewalVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsRegistryAuthRenewal() (*RegistryAuthRenewal, error) {
	return parseSettingAs[*RegistryAuthRenewal](s)
}

func (s *Setting) MustAsRegistryAuthRenewal() *RegistryAuthRenewal {
	return gofn.Must(s.AsRegistryAuthRenewal())
}
