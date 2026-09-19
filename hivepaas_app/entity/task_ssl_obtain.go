package entity

// TaskSSLObtainArgs carries what obtaining a certificate needs.
//
// The certificate setting is named rather than carried: it is read fresh when
// the task runs, which may be minutes later after a retry, and what matters then
// is what the setting says at that moment. AppID is the app whose domain started
// this, so its routing can be applied again once there is something to serve it
// with - the certificate is of no use to it until traefik is told.
type TaskSSLObtainArgs struct {
	SettingID string `json:"settingId"`
	AppID     string `json:"appId,omitempty"`
	// Domains are the addresses that were waiting on this certificate, kept for
	// the record: what the certificate covers is the certificate's own business.
	Domains []string `json:"domains,omitempty"`
}

type TaskSSLObtainOutput struct {
	Domain   string `json:"domain"`
	Obtained bool   `json:"obtained"`
	Applied  bool   `json:"applied,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (t *Task) ArgsAsSSLObtain() (*TaskSSLObtainArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSSLObtainArgs { return &TaskSSLObtainArgs{} })
}

func (t *Task) OutputAsSSLObtain() (*TaskSSLObtainOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSSLObtainOutput { return &TaskSSLObtainOutput{} })
}
