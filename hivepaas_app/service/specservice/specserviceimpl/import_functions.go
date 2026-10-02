package specserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// appNature is what an app is once imported: its kind's category, and its
// deployment source - what the bundle writes of them, else what the app has.
type appNature struct {
	category base.AppCategory
	source   *entity.AppDeploymentSettings
}

func (n appNature) isFunction() bool {
	return n.category == base.AppCategoryFunction
}

// checkFunctions holds a function to what creating one checks, whichever way
// the import writes it: its kind and its source agree, an app keeps its kind,
// its source is valid, and a cluster of several nodes has a registry to push
// its image to. What fails is blocked: nothing of the import is written. The
// functions found are remembered, for the writer to pin their routing to their
// runtime's port.
func (p *planner) checkFunctions(ctx context.Context) error {
	p.functionApps = map[string]bool{}
	var multiNode *bool
	for _, node := range p.writingApps() {
		if err := p.checkFunction(ctx, node, &multiNode); err != nil {
			return err
		}
	}
	return nil
}

// checkFunction checks one app the import writes; multiNode is asked of the
// cluster once, for the first function whose source is written.
func (p *planner) checkFunction(ctx context.Context, node *specmodel.PlanNode, multiNode **bool) error {
	writesKind := writesBlock(node, string(specmodel.BlockSettingsKind))
	writesSource := writesBlock(node, string(specmodel.BlockDeploymentSource))

	current, err := natureOf(p.currentApp(node))
	if err != nil {
		return err
	}
	bundled, err := natureOf(p.apps[node.Path].doc)
	if err != nil {
		return err
	}
	after := current
	if writesKind {
		after.category = bundled.category
	}
	if writesSource {
		after.source = bundled.source
	}
	p.functionApps[node.Path] = after.isFunction()
	if !writesKind && !writesSource {
		return nil
	}

	if node.Action == specmodel.ActionUpdate && writesKind && current.isFunction() != after.isFunction() {
		p.block(node, specmodel.CodeAppKindChanged, nil,
			"nothing can be imported: an app does not change its kind - a function stays one, and an app "+
				"does not become one")
		return nil
	}
	methodIsFunction := after.source != nil && after.source.ActiveMethod == base.DeploymentMethodFunction
	hasFunctionSource := after.source != nil && after.source.FunctionSource != nil
	if after.isFunction() != methodIsFunction || methodIsFunction != hasFunctionSource {
		p.block(node, specmodel.CodeFunctionKindMismatch, nil,
			"nothing can be imported: a function is an app of kind function deployed from a function's source, "+
				"and no other app has either")
		return nil
	}
	if !writesSource || !hasFunctionSource {
		return nil
	}

	source := after.source.FunctionSource
	normalizeFunctionSource(source)
	if problems := functionSourceProblems(source); len(problems) > 0 {
		p.block(node, specmodel.CodeFunctionSourceInvalid, map[string]any{"problems": problems},
			"nothing can be imported: the function's source is not valid")
		return nil
	}
	if *multiNode == nil {
		multi, err := p.s.clusterService.IsMultiNode(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
		*multiNode = &multi
	}
	if **multiNode && source.PushToRegistry.ID == "" {
		p.block(node, specmodel.CodeFunctionBuildSource, nil,
			"nothing can be imported: a cluster of several nodes pulls the function's image from a registry, "+
				"and the function names none to push it to")
	}
	return nil
}

func (p *planner) block(node *specmodel.PlanNode, code string, detail map[string]any, action string) {
	node.Issues = append(node.Issues, specmodel.Issue{
		Severity: specmodel.SeverityBlocked, Code: code, Path: node.Path, Detail: detail, Action: action,
	})
}

// natureOf reads an app document's kind and deployment source; nothing for no
// document, or a block a newer HivePaaS wrote, which SETTING_VERSION_NEWER
// blocks already.
func natureOf(doc *specmodel.AppDoc) (appNature, error) {
	var nature appNature
	if doc == nil {
		return nature, nil
	}
	if body, found := doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppKind)]; found {
		data, err := decodeNatureBlock(specmodel.BlockSettingsKind, base.SettingTypeAppKind, body)
		if err != nil {
			return nature, err
		}
		if kind, ok := data.(*entity.AppKindSettings); ok {
			nature.category = kind.Category
		}
	}
	if doc.Deployment != nil && doc.Deployment.Source != nil {
		data, err := decodeNatureBlock(specmodel.BlockDeploymentSource, base.SettingTypeAppDeployment,
			doc.Deployment.Source)
		if err != nil {
			return nature, err
		}
		nature.source, _ = data.(*entity.AppDeploymentSettings)
	}
	return nature, nil
}

func decodeNatureBlock(block specmodel.Block, typ base.SettingType, body any) (entity.SettingData, error) {
	var externals []*specmodel.ExternalRef
	_, data, err := decodeImportedSetting(block, typ, "", withExternalPlaceholders(body, &externals))
	if errors.Is(err, hperrors.ErrDataVerNewerThanSystemVer) {
		return nil, nil
	}
	return data, err
}
