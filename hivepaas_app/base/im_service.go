package base

type IMServiceKind string

const (
	IMServiceKindSlack    IMServiceKind = "slack"
	IMServiceKindDiscord  IMServiceKind = "discord"
	IMServiceKindTelegram IMServiceKind = "telegram"
)

var (
	AllIMServiceKinds = []IMServiceKind{IMServiceKindSlack, IMServiceKindDiscord, IMServiceKindTelegram}
)
