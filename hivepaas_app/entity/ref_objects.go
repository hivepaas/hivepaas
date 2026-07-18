package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func NewRefObjects() *RefObjects {
	return &RefObjects{
		RefSettings: make(map[string]*Setting),
		RefApps:     make(map[string]*App),
		RefUsers:    make(map[string]*User),
	}
}

type RefObjects struct {
	RefSettings map[string]*Setting
	RefApps     map[string]*App
	RefUsers    map[string]*User
}

func (r *RefObjects) AddRefObjects(refObjects *RefObjects) {
	if refObjects == nil {
		return
	}

	if r.RefSettings == nil {
		r.RefSettings = make(map[string]*Setting, len(refObjects.RefSettings))
	}
	for _, refSetting := range refObjects.RefSettings {
		r.RefSettings[refSetting.ID] = refSetting
	}

	if r.RefApps == nil {
		r.RefApps = make(map[string]*App, len(refObjects.RefApps))
	}
	for _, refApp := range refObjects.RefApps {
		r.RefApps[refApp.ID] = refApp
	}

	if r.RefUsers == nil {
		r.RefUsers = make(map[string]*User, len(refObjects.RefUsers))
	}
	for _, refUser := range refObjects.RefUsers {
		r.RefUsers[refUser.ID] = refUser
	}
}

type RefObjectIDs struct {
	RefSettingIDs []string
	RefAppIDs     []string
	RefUserIDs    []string
}

func (r *RefObjectIDs) HasData() bool {
	return len(r.RefSettingIDs) > 0 || len(r.RefAppIDs) > 0 || len(r.RefUserIDs) > 0
}

func (r *RefObjectIDs) AddRefIDs(refIDs *RefObjectIDs) {
	if refIDs == nil {
		return
	}
	r.RefSettingIDs = append(r.RefSettingIDs, refIDs.RefSettingIDs...)
	r.RefAppIDs = append(r.RefAppIDs, refIDs.RefAppIDs...)
	r.RefUserIDs = append(r.RefUserIDs, refIDs.RefUserIDs...)
}

func (r *RefObjectIDs) GetRecursiveRefObjectIDs(refObjects *RefObjects) *RefObjectIDs {
	newRefIDs := &RefObjectIDs{}
	for _, setting := range refObjects.RefSettings {
		newRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	res := &RefObjectIDs{}
	for _, settingID := range newRefIDs.RefSettingIDs {
		if !gofn.Contain(r.RefSettingIDs, settingID) {
			res.RefSettingIDs = append(res.RefSettingIDs, settingID)
		}
	}
	for _, appID := range newRefIDs.RefAppIDs {
		if !gofn.Contain(r.RefAppIDs, appID) {
			res.RefAppIDs = append(res.RefAppIDs, appID)
		}
	}
	for _, userID := range newRefIDs.RefUserIDs {
		if !gofn.Contain(r.RefUserIDs, userID) {
			res.RefUserIDs = append(res.RefUserIDs, userID)
		}
	}
	return res
}

func (r *RefObjectIDs) GetResourceLinks(srcType base.ResourceType, srcID string) []*ResLink {
	resLinks := make([]*ResLink, 0, len(r.RefSettingIDs)+len(r.RefAppIDs)+len(r.RefUserIDs))
	timeNow := timeutil.NowUTC()
	for _, refSettingID := range r.RefSettingIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeSetting,
			DstID:     refSettingID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refAppID := range r.RefAppIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeApp,
			DstID:     refAppID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refUserID := range r.RefUserIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeUser,
			DstID:     refUserID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	return resLinks
}
