package getstarteddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

// ChecklistResp is what a new installation still has to do, for the dashboard's
// Get started card. Each item's status is todo, obtaining (the certificate
// only), failed (the certificate only) or done.
type ChecklistResp struct {
	DashboardCert *ChecklistItemResp `json:"dashboardCert"`
	TwoFactor     *ChecklistItemResp `json:"twoFactor"`
	GithubApp     *ChecklistItemResp `json:"githubApp"`
}

type ChecklistItemResp struct {
	Status string `json:"status"`
	// Domain is the dashboard certificate's: the name it is for.
	Domain string `json:"domain,omitempty"`
	// Error is why the last attempt at the dashboard's certificate failed.
	Error string `json:"error,omitempty"`
}

func TransformChecklist(checklist *getstartedservice.Checklist) *ChecklistResp {
	if checklist == nil {
		return nil
	}
	return &ChecklistResp{
		DashboardCert: TransformChecklistItem(&checklist.DashboardCert),
		TwoFactor:     TransformChecklistItem(&checklist.TwoFactor),
		GithubApp:     TransformChecklistItem(&checklist.GithubApp),
	}
}

func TransformChecklistItem(item *getstartedservice.Item) *ChecklistItemResp {
	return &ChecklistItemResp{
		Status: string(item.Status),
		Domain: item.Domain,
		Error:  item.Error,
	}
}
