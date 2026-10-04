package specuc

import (
	"context"
	"errors"
	"strings"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// ValidateCompose answers what creating a project from a compose file would
// do, writing nothing: each service as the app it becomes, what the file needs
// that the request lacks, and the import's plan.
func (uc *UC) ValidateCompose(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ValidateComposeReq,
) (*specdto.ValidateComposeResp, error) {
	read, err := uc.readCompose(ctx, uc.db, auth, req, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := read.data()
	if read.converted.Bundle != nil {
		if data.Plan, err = uc.specService.PlanBundle(ctx, uc.db, read.plan); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &specdto.ValidateComposeResp{Data: data}, nil
}

// ApplyCompose creates what a validate of the same body planned, as an import
// applies its plan: the database in one transaction, recorded as
// compose-import, then the new apps' deployments.
func (uc *UC) ApplyCompose(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ApplyComposeReq,
) (_ *specdto.ApplyComposeResp, err error) {
	var applied *specservice.ApplyImportResp
	var read *composeRead
	committed := false
	cleanup := func() error {
		if applied == nil || applied.Cleanup == nil {
			return nil
		}
		cleanupErr := applied.Cleanup(context.WithoutCancel(ctx))
		applied = nil
		return hperrors.Wrap(cleanupErr)
	}
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.NewPanic(rec))
		}
		if err != nil && !committed {
			err = errors.Join(err, cleanup())
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		if cleanupErr := cleanup(); cleanupErr != nil {
			return hperrors.Wrap(cleanupErr)
		}
		var txErr error
		read, applied, txErr = uc.applyComposeInTx(ctx, db, auth, req)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true

	after := uc.afterImport(ctx, uc.db, applied)
	resp := &specdto.ApplyComposeResp{Meta: after.Meta, Data: &specdto.ApplyComposeData{
		Plan: after.Data.Plan, Deployments: after.Data.Deployments,
	}}
	if project, findErr := uc.projectRepo.GetByKey(ctx, uc.db, read.project.Key); findErr == nil {
		resp.Data.Project = &basedto.ObjectIDResp{ID: project.ID}
	}
	return resp, nil
}

// applyComposeInTx is the transaction's part: the file read again, the import
// applied and recorded. What it created in docker comes back with an error
// too, for the caller to remove.
func (uc *UC) applyComposeInTx(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *specdto.ApplyComposeReq,
) (*composeRead, *specservice.ApplyImportResp, error) {
	read, err := uc.readCompose(ctx, db, auth, &req.ValidateComposeReq, true)
	if err != nil {
		return nil, nil, err
	}
	if read.converted.Bundle == nil {
		return read, nil, hperrors.Wrap(hperrors.ErrSpecImportBlocked).
			WithExtraDetail("a required variable has no value")
	}
	applied, err := uc.specService.ApplyBundle(ctx, db, &specservice.ApplyBundleReq{
		PlanBundleReq: *read.plan, OperatorID: auth.User.ID, PlanHash: req.PlanHash, AcceptIssues: req.AcceptIssues,
	})
	if err != nil {
		return read, applied, hperrors.Wrap(err)
	}
	return read, applied, uc.recordComposeImport(ctx, db, auth, read, applied)
}

// composeRead is a compose request read: the project it names, what the file
// converts into, and the plan request it makes.
type composeRead struct {
	project   *specdto.ComposeProjectResp
	converted *composeservice.ConvertResp
	plan      *specservice.PlanBundleReq
}

func (r *composeRead) data() *specdto.ValidateComposeData {
	return &specdto.ValidateComposeData{
		Project: r.project, Services: r.converted.Services, Variables: r.converted.Variables,
		Needs: r.converted.Needs, Profiles: r.converted.Profiles,
	}
}

// readCompose names the project as creating one does, converts the file, and
// makes the plan request with the caller's gates. The values of the file are
// the request's own, so mounting them reveals nothing stored: no reveal gate.
func (uc *UC) readCompose(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *specdto.ValidateComposeReq,
	record bool,
) (*composeRead, error) {
	name := gofn.Coalesce(req.Project.Name, composeFileName(req.Compose))
	if name == "" {
		return nil, hperrors.NewValidationErrors(vld.Validate(vld.Must(false).OnError(
			vld.SetField("project.name", nil), vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"))))
	}
	key, err := uc.projectService.CheckNewProjectName(ctx, db, name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	project := &specdto.ComposeProjectResp{Name: name, Key: key, Env: req.Project.Env,
		EnvKey: projecthelper.CalcProjectEnvKey(req.Project.Env)}

	gates := uc.importReq(auth, &specdto.ValidateImportReq{
		Scope: entity.NewObjectScopeGlobal(), Selection: req.Selection,
		Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate, DeployCreated: req.Deploy},
	}, record)
	gates.MayMountSecrets, gates.AuthorizeSecrets = nil, nil
	mayWriteCluster, err := gates.MayWriteCluster(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	convertReq := uc.convertReq(req, project, mayWriteCluster, gates.AllowPrivilegedApps && gates.Admin)
	convertReq.OwnerID = auth.User.ID
	converted, err := uc.composeService.Convert(ctx, convertReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	project.FileName = converted.FileName
	return &composeRead{project: project, converted: converted, plan: &specservice.PlanBundleReq{
		ValidateImportReq: *gates, Doc: converted.Bundle, Issues: converted.Issues,
	}}, nil
}

func (uc *UC) convertReq(
	req *specdto.ValidateComposeReq,
	project *specdto.ComposeProjectResp,
	mayWriteCluster, mayBindHost bool,
) *composeservice.ConvertReq {
	out := &composeservice.ConvertReq{
		Compose: req.Compose, DotEnv: req.DotEnv, Files: req.Files, Profiles: req.Profiles,
		ProjectKey: project.Key, ProjectName: project.Name, EnvKey: project.EnvKey, EnvName: project.Env,
		NetworkName:     networkservice.ProjectNetworkName(&entity.Project{Key: project.Key}, project.Env),
		MayWriteCluster: mayWriteCluster, MayBindHost: mayBindHost,
		Variables: map[string]*composeservice.VariableReq{},
		Services:  map[string]*composeservice.ServiceReq{},
	}
	if cfg := config.Current(); cfg != nil {
		out.RootDomain = cfg.RootDomain
	}
	if req.Volume.ID != "" {
		out.Volume = &specmodel.ExternalRef{Type: string(base.SettingTypeClusterVolume), ID: req.Volume.ID}
	}
	for name, v := range req.Variables {
		if v != nil {
			out.Variables[name] = &composeservice.VariableReq{Value: v.Value, Secret: v.Secret}
		}
	}
	for name, svc := range req.Services {
		if svc == nil {
			continue
		}
		choice := &composeservice.ServiceReq{Image: svc.Image}
		for _, port := range svc.Ports {
			if port != nil {
				choice.Ports = append(choice.Ports, &composeservice.PortReq{Published: port.Published,
					Target: port.Target, Protocol: port.Protocol, As: port.As, Domain: port.Domain})
			}
		}
		out.Services[name] = choice
	}
	return out
}

// composeFileName is the project's name in the file, `name:`: one that holds a
// variable is not a name until it is read, and is left to the request.
func composeFileName(compose string) string {
	var top struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal([]byte(compose), &top); err != nil || strings.Contains(top.Name, "$") {
		return ""
	}
	return strings.TrimSpace(top.Name)
}

// recordComposeImport records a project created from a compose file in its
// transaction: what was read into it, where, and what the import did, counted.
func (uc *UC) recordComposeImport(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	read *composeRead,
	applied *specservice.ApplyImportResp,
) error {
	plan := applied.Plan
	detail := auditdetail.New().
		Set("project", read.project.Name).
		Set("env", read.project.Env).
		Set("digest", plan.Bundle.Digest).
		Set("deployCreated", read.plan.Options.DeployCreated).
		Set("exclude", read.plan.Selection.Exclude).
		Set("summary", plan.Summary)
	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeComposeImport,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIAction,
		Auth:    auth,
		ResType: base.ResourceTypeProject,
		ResName: read.project.Name,
		Detail:  detail.String(),
	})
	return hperrors.Wrap(err)
}
