package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentNotificationVersion = 1
)

var _ = registerSettingParser(base.SettingTypeNotification, &notificationParser{})

type notificationParser struct {
}

func (s *notificationParser) New() SettingData {
	return &Notification{}
}

type Notification struct {
	ViaEmail        *NotificationViaEmail    `json:"viaEmail,omitempty"`
	ViaSlack        *NotificationViaSlack    `json:"viaSlack,omitempty"`
	ViaDiscord      *NotificationViaDiscord  `json:"viaDiscord,omitempty"`
	ViaTelegram     *NotificationViaTelegram `json:"viaTelegram,omitempty"`
	MinSendInterval timeutil.Duration        `json:"minSendInterval,omitempty"`
}

func (s *Notification) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.ViaEmail.GetRefSettingIDs(), s.ViaSlack.GetRefSettingIDs(),
			s.ViaDiscord.GetRefSettingIDs(), s.ViaTelegram.GetRefSettingIDs()),
	}
}

func (s *Notification) HasNotificationViaEmail() bool {
	return s.ViaEmail != nil
}

func (s *Notification) HasNotificationViaSlack() bool {
	return s.ViaSlack != nil
}

func (s *Notification) HasNotificationViaDiscord() bool {
	return s.ViaDiscord != nil
}

func (s *Notification) HasNotificationViaTelegram() bool {
	return s.ViaTelegram != nil
}

type NotificationViaEmail struct {
	Enabled          bool     `json:"enabled"`
	UseDefault       bool     `json:"useDefault"` // If true, use the default email account in current scope
	Sender           ObjectID `json:"sender,omitzero"`
	ToProjectMembers bool     `json:"toProjectMembers,omitempty"`
	ToProjectOwners  bool     `json:"toProjectOwners,omitempty"`
	ToAllAdmins      bool     `json:"toAllAdmins,omitempty"`
	ToAddresses      []string `json:"toAddresses,omitempty"`
}

func (s *NotificationViaEmail) GetRefSettingIDs() (res []string) {
	if s == nil || !s.Enabled {
		return nil
	}
	if s.Sender.ID != "" {
		res = append(res, s.Sender.ID)
	}
	return res
}

type NotificationViaSlack struct {
	Enabled    bool     `json:"enabled"`
	UseDefault bool     `json:"useDefault"` // If true, use the default slack webhook in current scope
	Webhook    ObjectID `json:"webhook,omitzero"`
}

func (s *NotificationViaSlack) GetRefSettingIDs() (res []string) {
	if s == nil || !s.Enabled {
		return nil
	}
	if s.Webhook.ID != "" {
		res = append(res, s.Webhook.ID)
	}
	return res
}

type NotificationViaDiscord struct {
	Enabled    bool     `json:"enabled"`
	UseDefault bool     `json:"useDefault"` // If true, use the default discord webhook in current scope
	Webhook    ObjectID `json:"webhook,omitzero"`
}

func (s *NotificationViaDiscord) GetRefSettingIDs() (res []string) {
	if s == nil || !s.Enabled {
		return nil
	}
	if s.Webhook.ID != "" {
		res = append(res, s.Webhook.ID)
	}
	return res
}

type NotificationViaTelegram struct {
	Enabled    bool     `json:"enabled"`
	UseDefault bool     `json:"useDefault"` // If true, use the default telegram config in current scope
	Setting    ObjectID `json:"setting,omitzero"`
}

func (s *NotificationViaTelegram) GetRefSettingIDs() (res []string) {
	if s == nil || !s.Enabled {
		return nil
	}
	if s.Setting.ID != "" {
		res = append(res, s.Setting.ID)
	}
	return res
}

func (s *Notification) GetType() base.SettingType {
	return base.SettingTypeNotification
}

func (s *Notification) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Notification) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentNotificationVersion {
		return false, nil
	}
	if setting.Version > CurrentNotificationVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentNotificationVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Notification) Decrypt() error {
	return nil
}

func (s *Setting) AsNotification() (*Notification, error) {
	return parseSettingAs[*Notification](s)
}

func (s *Setting) MustAsNotification() *Notification {
	return gofn.Must(s.AsNotification())
}

func NewNotificationDefaultForScope(scope *base.ObjectScope) *Notification {
	notif := &Notification{
		ViaEmail: &NotificationViaEmail{
			Enabled:    true,
			UseDefault: true,
		},
		ViaSlack: &NotificationViaSlack{
			Enabled:    true,
			UseDefault: true,
		},
		ViaDiscord: &NotificationViaDiscord{
			Enabled:    true,
			UseDefault: true,
		},
		ViaTelegram: &NotificationViaTelegram{
			Enabled:    true,
			UseDefault: true,
		},
	}
	switch scope.ScopeType() {
	case base.ObjectScopeProject, base.ObjectScopeApp:
		notif.ViaEmail.ToProjectOwners = true
		notif.ViaEmail.ToProjectMembers = true
	case base.ObjectScopeUser:
		// Do nothing
	case base.ObjectScopeGlobal:
		notif.ViaEmail.ToAllAdmins = true
	}
	return notif
}
