package base

type IMServiceKind string

const (
	IMServiceKindSlack    IMServiceKind = "slack"
	IMServiceKindDiscord  IMServiceKind = "discord"
	IMServiceKindTelegram IMServiceKind = "telegram"
	IMServiceKindLark     IMServiceKind = "lark"
)

var (
	AllIMServiceKinds = []IMServiceKind{IMServiceKindSlack, IMServiceKindDiscord, IMServiceKindTelegram, IMServiceKindLark}
)
