package apptemplatedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// PreflightAppFromTemplateReq is the create request, asked about rather than
// carried out. It is the same body on purpose: a check of a different request
// than the one that follows checks nothing.
type PreflightAppFromTemplateReq struct {
	CreateAppFromTemplateReq

	// Lang is what the issues are worded in. They are the same refusals the
	// creation would raise, so they are translated the same way.
	Lang translation.Lang `json:"-"`
}

func NewPreflightAppFromTemplateReq() *PreflightAppFromTemplateReq {
	return &PreflightAppFromTemplateReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *PreflightAppFromTemplateReq) ModifyRequest() error {
	return req.CreateAppFromTemplateReq.ModifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *PreflightAppFromTemplateReq) Validate() hperrors.ValidationErrors {
	return req.CreateAppFromTemplateReq.Validate()
}

type PreflightAppFromTemplateResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *PreflightAppResult `json:"data"`
}

// PreflightAppResult is what the create dialog needs to know before it creates
// anything.
type PreflightAppResult struct {
	// Storage is every app of this request whose directory already holds
	// something. Empty is the ordinary case and means there is nothing to say.
	Storage []*PreflightStorageResult `json:"storage"`
	// StorageUnchecked is what could not be looked at - storage on a node that
	// could not be reached. It is reported because an empty Storage that means
	// "nothing was seen" reads exactly like one that means "there is nothing
	// there", and the second is the one that lets a deploy walk into old data.
	StorageUnchecked []*PreflightStorageResult `json:"storageUnchecked"`

	// Issues are the refusals the creation would raise: a domain already served,
	// a published port already held, permission the caller does not have. They
	// are reported together rather than one at a time, because a dialog that
	// says "and also" three times is three round trips through the same form.
	Issues []*PreflightIssueResult `json:"issues"`

	// Apps are the apps the request would create, in the order it creates them:
	// dependencies first, the app asked for last. An assistant shows them as the
	// plan of an install; nothing here is decided by anything but the request.
	Apps []*PreflightPlannedApp `json:"apps"`
}

// PreflightPlannedApp is one app a request would create, as it would be named.
type PreflightPlannedApp struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	// Kind is app for the app asked for, component for another process of the
	// same application, dependency for an app it needs.
	Kind string `json:"kind"`
	// Role is its role in the template - db, worker - empty for the app asked for.
	Role     string `json:"role,omitempty"`
	Template string `json:"template"`
	Version  string `json:"version,omitempty"`
	Image    string `json:"image"`
}

type PreflightIssueResult struct {
	// Code is the error the creation would fail with, for a screen that wants to
	// act on a particular one.
	Code string `json:"code"`
	// Detail is the sentence that refusal carries, in the caller's language.
	Detail string `json:"detail"`
}

type PreflightStorageResult struct {
	// App is the name the app will be created with, and AppKey what it will
	// answer to - which is also what its directory is called.
	App    string `json:"app"`
	AppKey string `json:"appKey"`
	// IsDatabase is the one distinction the warning makes: a database started on
	// somebody else's data keeps the password that data was created with, and the
	// one being generated now will not open it.
	IsDatabase bool `json:"isDatabase"`
	Volume     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"volume"`
	// Path is the directory inside the volume, so an operator can go and look.
	Path string `json:"path"`
}
