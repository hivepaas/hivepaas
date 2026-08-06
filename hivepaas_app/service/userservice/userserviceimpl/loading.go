package userserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
)

func (s *service) LoadUser(
	ctx context.Context,
	db database.IDB,
	userID string,
	errorIfUnavailable bool,
) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, db, userID,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if errorIfUnavailable {
		if err = s.checkUserAvailable(user); err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	return user, nil
}

func (s *service) LoadUsers(
	ctx context.Context,
	db database.IDB,
	userIDs []string,
	errorIfUnavailable bool,
) (map[string]*entity.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	userIDs = gofn.ToSet(userIDs)
	users, err := s.userRepo.ListByIDs(ctx, db, userIDs,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	userMap := entityutil.SliceToIDMap(users)

	for _, userID := range userIDs {
		if _, ok := userMap[userID]; !ok {
			return nil, apperrors.Wrap(apperrors.ErrUserNotFound).WithParam("Name", userID)
		}
	}

	resultMap, err := s.collectAvailUsers(userMap, userIDs, errorIfUnavailable)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resultMap, nil
}

func (s *service) LoadUsersSkipMissing(
	ctx context.Context,
	db database.IDB,
	userIDs []string,
	errorIfUnavailable bool,
) (map[string]*entity.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	userIDs = gofn.ToSet(userIDs)
	users, err := s.userRepo.ListByIDs(ctx, db, userIDs,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	userMap := entityutil.SliceToIDMap(users)

	resultMap, err := s.collectAvailUsers(userMap, userIDs, errorIfUnavailable)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resultMap, nil
}

func (s *service) collectAvailUsers(
	userMap map[string]*entity.User,
	requiredKeys []string,
	errorIfUnavailable bool,
) (map[string]*entity.User, error) {
	resultMap := make(map[string]*entity.User, len(userMap))
	for _, userKey := range requiredKeys {
		user := userMap[userKey]
		err := s.checkUserAvailable(user)
		if err != nil {
			if errorIfUnavailable {
				return nil, err
			}
		}
		resultMap[userKey] = user
	}
	return resultMap, nil
}

func (s *service) checkUserAvailable(user *entity.User) error {
	if user == nil {
		return apperrors.NewNotFound("User")
	}
	if user.Status != base.UserStatusActive {
		return apperrors.Wrap(apperrors.ErrUserUnavailable).
			WithMsgLog("user '%s' is not active", user.Username)
	}
	if user.IsAccessExpired() {
		return apperrors.Wrap(apperrors.ErrUserUnavailable).
			WithMsgLog("user '%s' has access expired at: %v", user.Username, user.AccessExpireAt)
	}
	if user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
		return apperrors.Wrap(apperrors.ErrUserNotCompleteMFASetup).
			WithMsgLog("user '%s' hasn't completed the MFA setup", user.Username)
	}
	return nil
}

func (s *service) LoadUsersCustomConds(
	ctx context.Context,
	db database.IDB,
	errorIfUnavailable bool,
	conds ...bunex.SelectQueryOption,
) (map[string]*entity.User, error) {
	if len(conds) == 0 {
		return nil, apperrors.NewArgumentInvalid("conds").
			WithMsgLog("LoadUsersCustomConds requires at least one condition")
	}

	users, _, err := s.userRepo.List(ctx, db, nil, conds...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	userMap := entityutil.SliceToIDMap(users)
	userIDs := gofn.MapKeys(userMap)

	resultMap, err := s.collectAvailUsers(userMap, userIDs, errorIfUnavailable)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resultMap, nil
}
