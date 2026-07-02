package userserviceimpl

import (
	"context"
	"strings"

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
) (*entity.User, error) {
	userMap, err := s.LoadUsers(ctx, db, []string{userID}, true)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(userMap) == 0 {
		return nil, apperrors.NewNotFound("User")
	}
	return userMap[userID], nil
}

func (s *service) LoadUserEx(
	ctx context.Context,
	db database.IDB,
	userID string,
	errorIfUnavail bool,
) (*entity.User, error) {
	userMap, err := s.LoadUsers(ctx, db, []string{userID}, errorIfUnavail)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(userMap) == 0 {
		return nil, apperrors.NewNotFound("User")
	}
	return userMap[userID], nil
}

func (s *service) LoadUsers(
	ctx context.Context,
	db database.IDB,
	userIDs []string,
	errorIfUnavail bool,
) (map[string]*entity.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	userIDs = gofn.ToSet(userIDs)
	users, err := s.userRepo.ListByIDs(ctx, db, userIDs,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}
	userMap := entityutil.SliceToIDMap(users)

	resultMap, err := s.collectAvailUsers(userMap, userIDs, errorIfUnavail)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resultMap, nil
}

func (s *service) LoadUserByEmail(
	ctx context.Context,
	db database.IDB,
	email string,
) (*entity.User, error) {
	userMap, err := s.LoadUsersByEmails(ctx, db, []string{email}, true)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(userMap) == 0 {
		return nil, apperrors.NewNotFound("User")
	}
	return userMap[email], nil
}

func (s *service) LoadUsersByEmails(
	ctx context.Context,
	db database.IDB,
	emails []string,
	errorIfUnavail bool,
) (map[string]*entity.User, error) {
	if len(emails) == 0 {
		return nil, nil
	}

	lowercaseEmails := gofn.MapSlice(emails, strings.ToLower)
	users, err := s.userRepo.ListByEmails(ctx, db, lowercaseEmails,
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}
	userMap := make(map[string]*entity.User, len(users))
	for _, user := range users {
		userMap[user.Email] = user
	}

	resultMap, err := s.collectAvailUsers(userMap, lowercaseEmails, errorIfUnavail)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resultMap, nil
}

func (s *service) collectAvailUsers(
	userMap map[string]*entity.User,
	requiredKeys []string,
	errorIfUnavail bool,
) (map[string]*entity.User, error) {
	resultMap := make(map[string]*entity.User, len(userMap))
	for _, userKey := range requiredKeys {
		user := userMap[userKey]
		if user == nil {
			if errorIfUnavail {
				return nil, apperrors.NewNotFound("User").
					WithMsgLog("user '%s' not found", userKey)
			}
			continue
		}
		if user.Status != base.UserStatusActive {
			if errorIfUnavail {
				return nil, apperrors.New(apperrors.ErrUserUnavailable).
					WithMsgLog("user '%s' is not active", userKey)
			}
			continue
		}
		if user.IsAccessExpired() {
			if errorIfUnavail {
				return nil, apperrors.New(apperrors.ErrUserUnavailable).
					WithMsgLog("user '%s' has access expired at: %v", userKey, user.AccessExpireAt)
			}
			continue
		}
		if user.SecurityOption == base.UserSecurityPassword2FA && user.TotpSecret == "" {
			if errorIfUnavail {
				return nil, apperrors.New(apperrors.ErrUserNotCompleteMFASetup).
					WithMsgLog("user '%s' hasn't completed the MFA setup", userKey)
			}
			continue
		}
		resultMap[userKey] = user
	}

	return resultMap, nil
}

func (s *service) LoadUsersEx(
	ctx context.Context,
	db database.IDB,
	errorIfUnavail bool,
	loadOpts ...bunex.SelectQueryOption,
) (map[string]*entity.User, error) {
	users, _, err := s.userRepo.List(ctx, db, nil, loadOpts...)
	if err != nil {
		return nil, apperrors.New(err)
	}

	userMap := entityutil.SliceToIDMap(users)
	userIDs := gofn.MapKeys(userMap)

	resultMap, err := s.collectAvailUsers(userMap, userIDs, errorIfUnavail)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resultMap, nil
}
