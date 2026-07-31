package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func NewRefObjects() *RefObjects {
	return &RefObjects{
		RefSettings:    make(map[string]*Setting, 5), //nolint:mnd
		RefApps:        make(map[string]*App, 2),     //nolint:mnd
		RefProjects:    make(map[string]*Project),
		RefProjectEnvs: make(map[string]*ProjectEnv),
		RefUsers:       make(map[string]*User),
	}
}

type RefObjects struct {
	RefSettings    map[string]*Setting    `json:"settings"`
	RefApps        map[string]*App        `json:"apps"`
	RefProjects    map[string]*Project    `json:"projects"`
	RefProjectEnvs map[string]*ProjectEnv `json:"projectEnvs"`
	RefUsers       map[string]*User       `json:"users"`
}

func (r *RefObjects) AddRefObjects(refObjects *RefObjects) {
	if refObjects == nil {
		return
	}

	for _, refSetting := range refObjects.RefSettings {
		r.RefSettings[refSetting.ID] = refSetting
	}
	for _, refApp := range refObjects.RefApps {
		r.RefApps[refApp.ID] = refApp
	}
	for _, refProject := range refObjects.RefProjects {
		r.RefProjects[refProject.ID] = refProject
	}
	for _, refProjectEnv := range refObjects.RefProjectEnvs {
		r.RefProjectEnvs[refProjectEnv.ID] = refProjectEnv
	}
	for _, refUser := range refObjects.RefUsers {
		r.RefUsers[refUser.ID] = refUser
	}
}

type RefObjectIDs struct {
	RefSettingIDs    []string `json:"settingIds"`
	RefAppIDs        []string `json:"appIds"`
	RefProjectIDs    []string `json:"projectIds"`
	RefProjectEnvIDs []string `json:"projectEnvIds"`
	RefUserIDs       []string `json:"userIds"`
}

func (r *RefObjectIDs) HasData() bool {
	return len(r.RefSettingIDs) > 0 || len(r.RefAppIDs) > 0 ||
		len(r.RefProjectIDs) > 0 || len(r.RefProjectEnvIDs) > 0 || len(r.RefUserIDs) > 0
}

func (r *RefObjectIDs) AddRefIDs(refIDs *RefObjectIDs) {
	if refIDs == nil {
		return
	}
	r.RefSettingIDs = append(r.RefSettingIDs, refIDs.RefSettingIDs...)
	r.RefAppIDs = append(r.RefAppIDs, refIDs.RefAppIDs...)
	r.RefProjectIDs = append(r.RefProjectIDs, refIDs.RefProjectIDs...)
	r.RefProjectEnvIDs = append(r.RefProjectEnvIDs, refIDs.RefProjectEnvIDs...)
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
	for _, projectID := range newRefIDs.RefProjectIDs {
		if !gofn.Contain(r.RefProjectIDs, projectID) {
			res.RefProjectIDs = append(res.RefProjectIDs, projectID)
		}
	}
	for _, projectEnvID := range newRefIDs.RefProjectEnvIDs {
		if !gofn.Contain(r.RefProjectEnvIDs, projectEnvID) {
			res.RefProjectEnvIDs = append(res.RefProjectEnvIDs, projectEnvID)
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
	resLinks := make([]*ResLink, 0, len(r.RefSettingIDs)+len(r.RefAppIDs)+len(r.RefProjectIDs)+
		len(r.RefProjectEnvIDs)+len(r.RefUserIDs))
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
	for _, refProjectID := range r.RefProjectIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeProject,
			DstID:     refProjectID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refProjectEnvID := range r.RefProjectEnvIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeProjectEnv,
			DstID:     refProjectEnvID,
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
