package entity

type TaskSSLRenewalArgs struct {
	TargetSSLs ObjectIDSlice `json:"targetSSLs"`
}

type TaskSSLRenewalOutput struct {
	RenewedSSLs          ObjectIDSlice `json:"renewedSSLs,omitempty"`
	ExpiringNotifiedSSLs ObjectIDSlice `json:"expiringNotifiedSSLs,omitempty"`
}

func (t *Task) ArgsAsSSLRenewal() (*TaskSSLRenewalArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSSLRenewalArgs { return &TaskSSLRenewalArgs{} })
}

func (t *Task) OutputAsSSLRenewal() (*TaskSSLRenewalOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSSLRenewalOutput { return &TaskSSLRenewalOutput{} })
}
