package specserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// ownerField names a project's owner, as a change and in an issue's detail.
const ownerField = "owner"

// planOwner decides who owns a project import writes, and returns the change
// that makes: "owner", or nothing. Users do not travel, so the bundle's owner is
// looked for here - by id, which finds them on the installation that exported
// the bundle even if their email changed since, and then by email, which finds
// the same person under another id elsewhere.
//
// A project created for an owner nobody here is goes to the operator importing
// it. An existing one keeps its owner then: handing a project to whoever
// happens to import it is a change nobody asked for. An existing project given
// another owner takes the gate project update applies.
func (p *planner) planOwner(
	ctx context.Context,
	node *specmodel.PlanNode,
	doc *specmodel.ProjectDoc,
	target *entity.Project,
) ([]string, error) {
	owner, err := p.resolveOwner(ctx, doc.Owner)
	if err != nil {
		return nil, err
	}
	if owner == nil {
		action := "the operator importing owns the project"
		if target != nil {
			action = "the project keeps its owner"
		}
		node.Notes = append(node.Notes, specmodel.Issue{
			Code: specmodel.CodeOwnerNotFound, Path: node.Path, Detail: ownerDetail(doc.Owner), Action: action,
		})
		return nil, nil
	}
	p.owners[node.Path] = owner.ID
	if target == nil || owner.ID == target.OwnerID || p.req.Options.Existing == specmodel.ExistingKeep {
		return nil, nil
	}
	if p.req.MayChangeOwner != nil {
		allowed, err := p.req.MayChangeOwner(ctx, target)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if !allowed {
			node.Issues = append(node.Issues, specmodel.Issue{
				Severity: specmodel.SeverityFixable, Code: specmodel.CodeOwnerNotPermitted, Path: node.Path,
				Detail: ownerDetail(doc.Owner),
				Action: "the project keeps its owner: changing it needs Write permission on the Project module",
			})
			return nil, nil
		}
	}
	return []string{ownerField}, nil
}

// resolveOwner is the active user a bundle's owner is here, or nil. A user who
// is disabled or has not accepted an invitation cannot own a project, as
// project update requires, and the search goes on past them.
func (p *planner) resolveOwner(ctx context.Context, owner *specmodel.ProjectOwner) (*entity.User, error) {
	if owner == nil {
		return nil, nil
	}
	for _, lookup := range []struct {
		key string
		get func() (*entity.User, error)
	}{
		{owner.ID, func() (*entity.User, error) { return p.s.userRepo.GetByID(ctx, p.db, owner.ID) }},
		{owner.Email, func() (*entity.User, error) { return p.s.userRepo.GetByEmail(ctx, p.db, owner.Email) }},
	} {
		if lookup.key == "" {
			continue
		}
		user, err := lookup.get()
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if user != nil && user.Status == base.UserStatusActive {
			return user, nil
		}
	}
	return nil, nil
}

// ownerDetail names the bundle's owner the way a person knows them.
func ownerDetail(owner *specmodel.ProjectOwner) map[string]any {
	switch {
	case owner == nil:
		return nil
	case owner.Email != "":
		return map[string]any{ownerField: owner.Email}
	default:
		return map[string]any{ownerField: owner.ID}
	}
}
