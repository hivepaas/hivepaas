package useruc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/emailservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

func (uc *UC) InviteUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.InviteUserReq,
) (*userdto.InviteUserResp, error) {
	inviteData := &userInviteData{}
	err := uc.loadUserInviteData(ctx, uc.db, req, inviteData)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserInviteData(req, inviteData, persistingData)

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err := uc.authorizeAccessChanges(ctx, db, auth, inviteData.User,
			accessResourceTypesToReplace(true, true), persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return uc.userService.PersistUserData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if req.SendInviteEmail {
		err = uc.emailService.SendMailUserInvite(ctx, uc.db, &emailservice.EmailDataUserInvite{
			BaseTemplateData: emailservice.BaseTemplateData{
				Email:      inviteData.SystemEmail,
				Recipients: []string{inviteData.User.Email},
				Subject:    "You’ve been invited to join HivePaaS",
			},
			InviterName:    gofn.Coalesce(auth.User.FullName, auth.User.Username),
			UserSignupLink: inviteData.InviteLink,
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return &userdto.InviteUserResp{
		Data: &userdto.InviteUserDataResp{
			InviteLink: inviteData.InviteLink,
		},
	}, nil
}

type userInviteData struct {
	User        *entity.User
	InviteLink  string
	SystemEmail *entity.Email
}

func (uc *UC) loadUserInviteData(
	ctx context.Context,
	db database.IDB,
	req *userdto.InviteUserReq,
	data *userInviteData,
) error {
	// A pending user being re-invited has grants this invite replaces.
	user, err := uc.userRepo.GetByEmail(ctx, db, req.Email,
		bunex.SelectRelation("Accesses"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if user != nil && user.Status != base.UserStatusPending {
		return hperrors.NewAlreadyExist("User").
			WithMsgLog("user '%s' already exists", req.Email)
	}

	if req.SendInviteEmail {
		emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		email, err := emailSetting.AsEmail()
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.SystemEmail = email
	}

	// Calculate username from the email
	username := strings.SplitN(req.Email, "@", 2)[0] //nolint:mnd
	// If the username is not available, append some random chars
	conflictUser, err := uc.userRepo.GetByUsername(ctx, db, username)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if conflictUser != nil {
		username += fmt.Sprintf("-%d", 1000+rand.Intn(9000)) //nolint
	}

	if user == nil {
		user = &entity.User{
			ID:        gofn.Must(ulid.NewStringULID()),
			Username:  username,
			Email:     req.Email,
			FullName:  username,
			CreatedAt: time.Now(),
		}
	}
	data.User = user

	// Generate invite token
	inviteToken, err := uc.userService.GenerateUserInviteToken(user.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.InviteLink = config.Current.DashboardUserSignupURL(inviteToken)

	return nil
}

func (uc *UC) preparePersistingUserInviteData(
	req *userdto.InviteUserReq,
	data *userInviteData,
	persistingData *userservice.PersistingUserData,
) {
	timeNow := timeutil.NowUTC()
	user := data.User

	user.Role = req.Role
	user.Status = base.UserStatusPending
	user.SecurityOption = req.SecurityOption
	user.AccessExpireAt = req.AccessExpireAt
	user.UpdatedAt = timeNow

	persistingData.UpsertingUsers = append(persistingData.UpsertingUsers, user)

	uc.preparePersistingUserModuleAccesses(user, req.ModuleAccesses, timeNow, persistingData)
	uc.preparePersistingUserProjectAccesses(user, req.ProjectAccesses, timeNow, persistingData)
}

func (uc *UC) preparePersistingUserModuleAccesses(
	user *entity.User,
	moduleReqs basedto.ModuleAccessSliceReq,
	timeNow time.Time,
	persistingData *userservice.PersistingUserData,
) {
	for _, moduleReq := range moduleReqs {
		persistingData.UpsertingAccesses = append(persistingData.UpsertingAccesses,
			&entity.ACLPermission{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeModule,
				ResourceID:   moduleReq.ID,
				Actions:      moduleReq.Access,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			})
	}
}

// preparePersistingUserProjectAccesses turns the requested grants into ACL rows.
// The request groups grants by project for symmetry with the GET response, but
// permissions are stored per project env only: the env IDs already carry their
// project, and userdto.validateEnvAccesses has checked their shape.
func (uc *UC) preparePersistingUserProjectAccesses(
	user *entity.User,
	projectReqs []*userdto.ProjectAccessReq,
	timeNow time.Time,
	persistingData *userservice.PersistingUserData,
) {
	for _, projectReq := range projectReqs {
		for _, envReq := range projectReq.EnvAccesses {
			persistingData.UpsertingAccesses = append(persistingData.UpsertingAccesses,
				&entity.ACLPermission{
					SubjectType:  base.SubjectTypeUser,
					SubjectID:    user.ID,
					ResourceType: base.ResourceTypeProjectEnv,
					ResourceID:   envReq.ID,
					Actions:      envReq.Access,
					CreatedAt:    timeNow,
					UpdatedAt:    timeNow,
				})
		}
	}
}
