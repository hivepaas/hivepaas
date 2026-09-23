package apptemplatedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// PreflightAppFromTemplateReq is the create request, asked about rather than
// carried out. It is the same body on purpose: a check of a different request
// than the one that follows checks nothing.
type PreflightAppFromTemplateReq struct {
	CreateAppFromTemplateReq
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
	Meta *basedto.Meta   `json:"meta"`
	Data *PreflightAppRe `json:"data"`
}

// PreflightAppRe is what the create dialog needs to know before it creates
// anything.
type PreflightAppRe struct {
	// Storage is every app of this request whose directory already holds
	// something. Empty is the ordinary case and means there is nothing to say.
	Storage []*PreflightStorageRes `json:"storage"`
	// StorageUnchecked is what could not be looked at - storage on a node that
	// could not be reached. It is reported because an empty Storage that means
	// "nothing was seen" reads exactly like one that means "there is nothing
	// there", and the second is the one that lets a deploy walk into old data.
	StorageUnchecked []*PreflightStorageRes `json:"storageUnchecked"`
}

type PreflightStorageRes struct {
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
