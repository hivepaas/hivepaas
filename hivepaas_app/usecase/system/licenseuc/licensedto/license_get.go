package licensedto

import (
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

type GetLicenseReq struct{}

func NewGetLicenseReq() *GetLicenseReq {
	return &GetLicenseReq{}
}

type GetLicenseResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *LicenseResp  `json:"data"`
}

// LicenseResp is the license in force, as System > License shows it: status is
// what the stored key amounts to on this installation now, plan is the plan in
// force (the key's, or Bee when the key puts nothing in force), and keyPlan is
// the plan the key was issued for, in force or not.
//
// The key's own fields - keyPlan, trial, licenseId, customerName, installations,
// issuedAt and expiresAt - are there whenever its signatures verified, also when
// it is not in force, as for another installation or past its grace, so the page
// can say which key it is. Of the customer only the name is sent: the key also
// carries an email, which never leaves the server.
//
// The enum fields carry no comment of their own: swag would then write them as
// an object holding the enum rather than as the enum.
type LicenseResp struct {
	Status  base.LicenseStatus `json:"status"`
	Plan    base.LicensePlan   `json:"plan"`
	KeyPlan base.LicensePlan   `json:"keyPlan,omitempty"`
	Trial   bool               `json:"trial"`
	// Development is a key signed by the development keys, which only a backend
	// in the development environment accepts.
	Development  bool   `json:"development"`
	LicenseID    string `json:"licenseId,omitempty"`
	CustomerName string `json:"customerName,omitempty"`
	// InstallationID is the ID a key names this installation by: its swarm's.
	// It is empty when docker cannot say.
	InstallationID string `json:"installationId"`
	// Installations are the IDs of the installations the key is for.
	Installations []string   `json:"installations,omitempty"`
	IssuedAt      *time.Time `json:"issuedAt,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	// GraceEndsAt is when a key for this installation stops holding: its grace
	// after it expires, or after it was moved if that comes first.
	GraceEndsAt *time.Time `json:"graceEndsAt,omitempty"`
	// MovedAt is when the key was moved to another installation.
	MovedAt *time.Time `json:"movedAt,omitempty"`
	// Reason says why the key is invalid.
	Reason string `json:"reason,omitempty"`
	// Limits has every kind a plan can limit, in the same order every time.
	Limits []*LicenseLimitResp `json:"limits"`
}

// LicenseLimitResp is one kind a plan limits: the most the plan in force allows,
// and how many there are.
type LicenseLimitResp struct {
	Limit base.LicenseLimit `json:"limit"`
	// Max is null when the plan in force does not limit the kind. Sign-in
	// providers are 0 or null: no paid kind can be added, or every one can.
	Max *int `json:"max"`
	// Used is how many there are, counted as a creation counts them. For sign-in
	// providers it is the paid ones active.
	Used int `json:"used"`
}

// TransformLicense writes the license in force and the usage into one response.
// It copies the customer's name and nothing else of the customer, so the email a
// key carries cannot reach the wire by a field added here later.
func TransformLicense(state *license.State, usage *licenseservice.Usage) *LicenseResp {
	resp := &LicenseResp{
		Status:         state.Status,
		Plan:           state.Plan,
		Development:    state.Development,
		InstallationID: state.InstallationID,
		GraceEndsAt:    optionalTime(state.GraceEndsAt),
		MovedAt:        optionalTime(state.MovedAt),
		Reason:         state.Reason,
	}
	if key := state.Key; key != nil {
		resp.KeyPlan = key.Plan
		resp.Trial = key.Trial
		resp.LicenseID = key.LicenseID
		resp.CustomerName = key.Customer.Name
		resp.Installations = slices.Clone(key.Installations)
		resp.IssuedAt = optionalTime(key.IssuedAt)
		resp.ExpiresAt = optionalTime(key.ExpiresAt)
	}

	resp.Limits = make([]*LicenseLimitResp, 0, len(base.AllLicenseLimits))
	for _, limit := range base.AllLicenseLimits {
		item := &LicenseLimitResp{Limit: limit}
		if maxCount, limited := state.Limits.Max(limit); limited {
			item.Max = &maxCount
		}
		if usage != nil {
			item.Used = usage.Used[limit]
		}
		resp.Limits = append(resp.Limits, item)
	}
	return resp
}

// optionalTime is nil for the zero time, which the response leaves out rather
// than sending as the year 1.
func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	t = t.UTC()
	return &t
}
