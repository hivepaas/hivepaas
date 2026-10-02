package entity

// TaskRegistryAuthRenewalArgs are a renewal's targets: the credentials whose
// services it renews; every Amazon ECR credential when empty.
type TaskRegistryAuthRenewalArgs struct {
	TargetAuths ObjectIDSlice `json:"targetAuths"`
}

// TaskRegistryAuthRenewalOutput is what a renewal did: the credentials it got a
// token for, the services it handed one to, and what failed - kept for the
// credential's page, and retried by the next run.
type TaskRegistryAuthRenewalOutput struct {
	Auths    ObjectIDSlice                 `json:"auths,omitempty"`
	Services []*RegistryAuthRenewalService `json:"services,omitempty"`
	Failures []*RegistryAuthRenewalFailure `json:"failures,omitempty"`
}

// RegistryAuthRenewalService is a service handed a credential's new token.
type RegistryAuthRenewalService struct {
	Auth      string `json:"auth"`
	App       string `json:"app"`
	ServiceID string `json:"serviceId"`
}

// RegistryAuthRenewalFailure is a credential no token was got for, App empty,
// or an app's service that could not be updated.
type RegistryAuthRenewalFailure struct {
	Auth  string `json:"auth"`
	App   string `json:"app,omitempty"`
	Error string `json:"error"`
}

func (t *Task) ArgsAsRegistryAuthRenewal() (*TaskRegistryAuthRenewalArgs, error) {
	return parseTaskArgsAs(t, func() *TaskRegistryAuthRenewalArgs { return &TaskRegistryAuthRenewalArgs{} })
}
