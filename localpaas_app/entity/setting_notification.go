package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
	ViaEmail        *NotificationViaEmail   `json:"viaEmail,omitempty"`
	ViaSlack        *NotificationViaSlack   `json:"viaSlack,omitempty"`
	ViaDiscord      *NotificationViaDiscord `json:"viaDiscord,omitempty"`
	MinSendInterval timeutil.Duration       `json:"minSendInterval,omitempty"`
}

func (s *Notification) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{
		RefSettingIDs: gofn.Flatten(s.ViaEmail.GetRefSettingIDs(), s.ViaSlack.GetRefSettingIDs(),
			s.ViaDiscord.GetRefSettingIDs()),
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

type NotificationViaEmail struct {
	Sender           ObjectID `json:"sender"`
	ToProjectMembers bool     `json:"toProjectMembers,omitempty"`
	ToProjectOwners  bool     `json:"toProjectOwners,omitempty"`
	ToAllAdmins      bool     `json:"toAllAdmins,omitempty"`
	ToAddresses      []string `json:"toAddresses,omitempty"`
}

func (s *NotificationViaEmail) GetRefSettingIDs() (res []string) {
	if s == nil {
		return res
	}
	res = append(res, s.Sender.ID)
	return res
}

type NotificationViaSlack struct {
	Webhook ObjectID `json:"webhook"`
}

func (s *NotificationViaSlack) GetRefSettingIDs() (res []string) {
	if s == nil {
		return res
	}
	res = append(res, s.Webhook.ID)
	return res
}

type NotificationViaDiscord struct {
	Webhook ObjectID `json:"webhook"`
}

func (s *NotificationViaDiscord) GetRefSettingIDs() (res []string) {
	if s == nil {
		return res
	}
	res = append(res, s.Webhook.ID)
	return res
}

func (s *Notification) GetType() base.SettingType {
	return base.SettingTypeNotification
}

func (s *Notification) MustDecrypt() *Notification {
	return s
}

func (s *Notification) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentNotificationVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentNotificationVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsNotification() (*Notification, error) {
	return parseSettingAs[*Notification](s)
}

func (s *Setting) MustAsNotification() *Notification {
	return gofn.Must(s.AsNotification())
}
