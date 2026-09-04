package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

// resourceKey identifies the resource an ACL row applies to.
type resourceKey struct {
	Type base.ResourceType
	ID   string
}

// changeKey identifies one subject's access to one resource. A single call may
// carry changes for several subjects at once.
type changeKey struct {
	SubjectID string
	Resource  resourceKey
}

// actorAccess is what the acting user is allowed to do, per resource.
type actorAccess map[resourceKey]base.AccessActions

// actionsOn returns the actions the actor holds on a resource. A project env the
// actor has no row for inherits the project-level actions, the same rule
// LoadProjectAccesses applies when resolving effective access.
func (a actorAccess) actionsOn(key resourceKey) base.AccessActions {
	if actions, ok := a[key]; ok {
		return actions
	}
	if key.Type == base.ResourceTypeProjectEnv {
		if projectID, _ := projecthelper.ParseProjectEnvID(key.ID); projectID != "" {
			return a[resourceKey{Type: base.ResourceTypeProject, ID: projectID}]
		}
	}
	return base.AccessActions{}
}

// AuthorizeAccessChanges checks the access changes the acting user wants to
// apply to another subject, and reports which of the subject's existing rows the
// operation may replace.
//
// A user must not be able to hand out - or take away - more than they hold
// themselves, so a non-admin may only flip the action bits they own: for every
// resource whose access changes, each action added or removed has to be one the
// actor has on that same resource. Actions left untouched are never checked, so
// editing a user an admin granted more to keeps working.
//
// The returned rows are the ones the caller may delete before writing `desired`.
// Everything else in `current` is out of the actor's reach and must be left
// alone, otherwise a wholesale replace would silently revoke grants the actor
// cannot even see. For an admin that is all of `current`.
//
// `desired` is the set of rows about to be written; a row with DeletedAt set is
// an explicit revocation.
func (p *manager) AuthorizeAccessChanges(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	desired []*entity.ACLPermission,
	current []*entity.ACLPermission,
) ([]*entity.ACLPermission, error) {
	// Admins have all privileges
	if auth.User.Role == base.UserRoleAdmin {
		return current, nil
	}

	actor, err := p.loadActorAccess(ctx, db, auth.User.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return authorizeAccessChanges(actor, desired, current)
}

// loadActorAccess reads the acting user's own grants.
func (p *manager) loadActorAccess(
	ctx context.Context,
	db database.IDB,
	userID string,
) (actorAccess, error) {
	perms, err := p.aclPermissionRepo.ListByUsers(ctx, db, []string{userID},
		bunex.SelectWhere("res_type IN (?)", bunex.List([]base.ResourceType{
			base.ResourceTypeModule,
			base.ResourceTypeProject,
			base.ResourceTypeProjectEnv,
		})),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	access := make(actorAccess, len(perms))
	for _, perm := range perms {
		access[resourceKey{Type: perm.ResourceType, ID: perm.ResourceID}] = perm.Actions
	}
	return access, nil
}

// authorizeAccessChanges is the pure part of AuthorizeAccessChanges.
func authorizeAccessChanges(actor actorAccess, desired, current []*entity.ACLPermission) (
	replaceable []*entity.ACLPermission, err error) {
	currentMap := make(map[changeKey]base.AccessActions, len(current))
	for _, perm := range current {
		if perm.DeletedAt.IsZero() {
			currentMap[keyOf(perm)] = perm.Actions
		}
	}

	for _, perm := range desired {
		wanted := perm.Actions
		if !perm.DeletedAt.IsZero() {
			wanted = base.AccessActions{} // the row is being revoked
		}
		key := keyOf(perm)
		if err := checkResourceChange(actor, key.Resource, currentMap[key], wanted); err != nil {
			return nil, err
		}
	}

	// A row may be replaced only when the actor could revoke every action it
	// carries: dropping it from the payload has to stay within their reach.
	for _, perm := range current {
		if !perm.DeletedAt.IsZero() {
			continue
		}
		if covers(actor.actionsOn(keyOf(perm).Resource), perm.Actions) {
			replaceable = append(replaceable, perm)
		}
	}
	return replaceable, nil
}

func keyOf(perm *entity.ACLPermission) changeKey {
	return changeKey{
		SubjectID: perm.SubjectID,
		Resource:  resourceKey{Type: perm.ResourceType, ID: perm.ResourceID},
	}
}

// covers reports whether `allowed` holds every action of `wanted`.
func covers(allowed, wanted base.AccessActions) bool {
	for _, action := range base.AllActionTypes {
		if wanted.Allows(action) && !allowed.Allows(action) {
			return false
		}
	}
	return true
}

// checkResourceChange rejects a change touching an action the actor does not hold.
func checkResourceChange(actor actorAccess, key resourceKey, current, desired base.AccessActions) error {
	if current.Equal(desired) {
		return nil
	}

	allowed := actor.actionsOn(key)
	for _, action := range base.AllActionTypes {
		if current.Allows(action) == desired.Allows(action) {
			continue // this action is not being changed
		}
		if !allowed.Allows(action) {
			return hperrors.Wrap(hperrors.ErrForbidden).
				WithMsgLog("you are not allowed to change the '%s' permission on %s '%s'",
					action, key.Type, key.ID)
		}
	}
	return nil
}
