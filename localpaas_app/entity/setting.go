package entity

import (
	"encoding/json"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
)

var (
	SettingUpsertingConflictCols = []string{"id"}
	SettingUpsertingUpdateCols   = []string{"object_id", "type", "kind", "status", "name", "data",
		"avail_in_projects", "is_default", "version", "update_ver",
		"updated_at", "expire_at", "deleted_at"}
)

type SettingParser interface {
	New() SettingData
}

var (
	settingParserMap = make(map[base.SettingType]SettingParser, 20) //nolint:mnd
)

//nolint:unparam
func registerSettingParser(typ base.SettingType, parser SettingParser) bool {
	settingParserMap[typ] = parser
	return true
}

type Setting struct {
	ID              string `bun:",pk"`
	ObjectID        string `bun:",nullzero"`
	Type            base.SettingType
	Kind            string `bun:",nullzero"`
	Status          base.SettingStatus
	Name            string `bun:",nullzero"`
	Data            string `bun:",nullzero"`
	AvailInProjects bool
	Default         bool `bun:"is_default"`
	Version         int
	UpdateVer       int

	CreatedAt time.Time `bun:",default:current_timestamp"`
	UpdatedAt time.Time `bun:",default:current_timestamp"`
	ExpireAt  time.Time `bun:",nullzero"`
	DeletedAt time.Time `bun:",soft_delete,nullzero"`

	BelongToUser    *User    `bun:"rel:belongs-to,join:object_id=id"`
	BelongToProject *Project `bun:"rel:belongs-to,join:object_id=id"`
	BelongToApp     *App     `bun:"rel:belongs-to,join:object_id=id"`
	Tasks           []*Task  `bun:"rel:has-many,join:id=target_id"`

	// NOTE: temporary fields
	parsedData      SettingData
	CurrentObjectID string `bun:"-"`
}

type SettingData interface {
	GetType() base.SettingType
	GetRefObjectIDs() *RefObjectIDs
	Migrate(setting *Setting) (hasChange bool, err error)
}

// GetID implements IDEntity interface
func (s *Setting) GetID() string {
	return s.ID
}

// GetName implements NamedEntity interface
func (s *Setting) GetName() string {
	return s.Name
}

// IsActive returns true if setting has status `active` and does not expire
func (s *Setting) IsActive() bool {
	return s.Status == base.SettingStatusActive && !s.IsDeleted() && !s.IsExpired()
}

func (s *Setting) IsDeleted() bool {
	return !s.DeletedAt.IsZero()
}

func (s *Setting) IsExpired() bool {
	return !s.ExpireAt.IsZero() && s.ExpireAt.Before(time.Now())
}

func (s *Setting) IsStatusDirty() bool {
	return s.Status == base.SettingStatusActive && s.IsExpired()
}

func (s *Setting) IsTypeIn(types ...base.SettingType) bool {
	return gofn.Contain(types, s.Type)
}

func (s *Setting) parseData(structPtr SettingData) error {
	if s == nil || s.Data == "" {
		return nil
	}
	err := json.Unmarshal(reflectutil.UnsafeStrToBytes(s.Data), structPtr)
	if err != nil {
		return apperrors.Wrap(err)
	}
	s.parsedData = structPtr
	return nil
}

func (s *Setting) SetData(data SettingData) error {
	if data.GetType() != s.Type {
		return apperrors.NewTypeInvalid()
	}
	b, err := json.Marshal(data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	s.Data = reflectutil.UnsafeBytesToStr(b)
	s.parsedData = data
	return nil
}

func (s *Setting) MustSetData(data SettingData) {
	gofn.Must1(s.SetData(data))
}

func (s *Setting) Parse() (SettingData, error) {
	return parseSettingAs[SettingData](s)
}

func (s *Setting) GetRefObjectIDs() (*RefObjectIDs, error) {
	settingData, err := s.Parse()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if settingData == nil {
		return nil, nil
	}
	return settingData.GetRefObjectIDs(), nil
}

func (s *Setting) MustGetRefObjectIDs() *RefObjectIDs {
	return gofn.Must(s.GetRefObjectIDs())
}

func parseSettingAs[T SettingData](s *Setting) (res T, err error) {
	if s.parsedData != nil {
		res, ok := s.parsedData.(T)
		if !ok {
			return res, apperrors.NewTypeInvalid()
		}
		return res, nil
	}
	if s.Data != "" {
		res = settingParserMap[s.Type].New().(T) //nolint:forcetypeassert
		if res.GetType() != s.Type {
			return res, apperrors.NewTypeInvalid()
		}
		if err := s.parseData(res); err != nil {
			return res, apperrors.Wrap(err)
		}
	}
	return res, nil
}

func (s *Setting) Migrate() (hasChange bool, err error) {
	settingData, err := s.Parse()
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	if settingData == nil {
		return false, nil
	}
	hasChange, err = settingData.Migrate(s)
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	return hasChange, nil
}
