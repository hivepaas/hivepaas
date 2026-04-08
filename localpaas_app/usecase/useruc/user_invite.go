package useruc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc/userdto"
)

func (uc *UC) InviteUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.InviteUserReq,
) (*userdto.InviteUserResp, error) {
	inviteData := &userInviteData{}
	err := uc.loadUserInviteData(ctx, uc.db, req, inviteData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserInviteData(req, inviteData, persistingData)

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		return uc.userService.PersistUserData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if req.SendInviteEmail {
		err = uc.emailService.SendMailUserInvite(ctx, uc.db, &emailservice.EmailDataUserInvite{
			Email:          inviteData.SystemEmail,
			Recipients:     []string{inviteData.User.Email},
			InviterName:    gofn.Coalesce(auth.User.FullName, auth.User.Username),
			UserSignupLink: inviteData.InviteLink,
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &userdto.InviteUserResp{
		Data: &userdto.InviteUserDataResp{
			InviteLink: inviteData.InviteLink,
		},
	}, nil
}

type userInviteData struct {
	User                  *entity.User
	RemoveCurrentAccesses bool
	InviteLink            string
	SystemEmail           *entity.Email
}

func (uc *UC) loadUserInviteData(
	ctx context.Context,
	db database.IDB,
	req *userdto.InviteUserReq,
	data *userInviteData,
) error {
	user, err := uc.userRepo.GetByEmail(ctx, db, req.Email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if user != nil && user.Status != base.UserStatusPending {
		return apperrors.NewAlreadyExist("User").
			WithMsgLog("user '%s' already exists", req.Email)
	}

	if req.SendInviteEmail {
		emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
		if err != nil {
			return apperrors.Wrap(err)
		}
		email, err := emailSetting.AsEmail()
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.SystemEmail = email
	}

	// Calculate username from the email
	username := strings.SplitN(req.Email, "@", 2)[0] //nolint:mnd
	// If the username is not available, append some random chars
	conflictUser, err := uc.userRepo.GetByUsername(ctx, db, username)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
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
	} else {
		// Remove all current accesses
		data.RemoveCurrentAccesses = true
	}
	data.User = user

	// Generate invite token
	inviteToken, err := uc.userService.GenerateUserInviteToken(user.ID)
	if err != nil {
		return apperrors.Wrap(err)
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

	if data.RemoveCurrentAccesses {
		persistingData.DeletingAccesses = append(persistingData.DeletingAccesses,
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeModule,
			},
			&base.PermissionResource{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeProject,
			},
		)
	}
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

func (uc *UC) preparePersistingUserProjectAccesses(
	user *entity.User,
	projectReqs basedto.ObjectAccessSliceReq,
	timeNow time.Time,
	persistingData *userservice.PersistingUserData,
) {
	for _, projectReq := range projectReqs {
		persistingData.UpsertingAccesses = append(persistingData.UpsertingAccesses,
			&entity.ACLPermission{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeProject,
				ResourceID:   projectReq.ID,
				Actions:      projectReq.Access,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			})
	}
}
