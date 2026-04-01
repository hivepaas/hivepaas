package userservice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/nanoid"
)

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	// 0644 grants read/write for owner, read-only for group/others
	photoDirFileMode = 0o755
)

func (s *userService) SaveUserPhoto(
	_ context.Context,
	user *entity.User,
	data []byte,
	fileExt string,
) error {
	dirPath := filepath.Join(config.Current.DataPathPhoto(), "user")
	err := os.MkdirAll(dirPath, photoDirFileMode)
	if err != nil {
		return fmt.Errorf("error creating user photo directory: %w", err)
	}

	// Remove current photo
	if user.Photo != "" {
		parts := strings.Split(user.Photo, "/")
		currentPhoto := parts[len(parts)-1]
		err = os.Remove(filepath.Join(dirPath, currentPhoto))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing old user photo: %w", err)
		}
	}

	if fileExt != "" && !strings.HasPrefix(fileExt, ".") {
		fileExt += "."
	}
	var fileName, filePath string
	i := 0
	for {
		fileName = nanoid.NewStandard16() + fileExt
		filePath = filepath.Join(dirPath, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		i++
		if i > 10 { //nolint:mnd
			return fmt.Errorf("error creating unique file name for user photo: %w",
				apperrors.ErrInternalServer)
		}
	}

	err = os.WriteFile(filePath, data, photoDirFileMode)
	if err != nil {
		return fmt.Errorf("error writing user photo: %w", err)
	}

	// Save the photo path
	user.Photo = filepath.Join(config.Current.HttpPathPhoto(), "user", fileName)
	return nil
}
