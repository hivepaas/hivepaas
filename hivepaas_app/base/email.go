package base

type EmailKind string

const (
	EmailKindSMTP EmailKind = "smtp"
	EmailKindHTTP EmailKind = "http"
)

var (
	AllEmailKinds = []EmailKind{EmailKindSMTP, EmailKindHTTP}
)
