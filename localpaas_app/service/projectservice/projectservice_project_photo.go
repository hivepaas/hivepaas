package projectservice

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

func (s *projectService) SaveProjectPhoto(
	_ context.Context,
	project *entity.Project,
	data []byte,
	fileExt string,
) error {
	dirPath := filepath.Join(config.Current.DataPathPhoto(), "project")
	err := os.MkdirAll(dirPath, photoDirFileMode)
	if err != nil {
		return fmt.Errorf("error creating project photo directory: %w", err)
	}

	// Remove current photo
	if project.Photo != "" {
		parts := strings.Split(project.Photo, "/")
		currentPhoto := parts[len(parts)-1]
		err = os.Remove(filepath.Join(dirPath, currentPhoto))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing old project photo: %w", err)
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
			return fmt.Errorf("error creating unique file name for project photo: %w",
				apperrors.ErrInternalServer)
		}
	}

	err = os.WriteFile(filePath, data, photoDirFileMode)
	if err != nil {
		return fmt.Errorf("error writing project photo: %w", err)
	}

	// Save the photo path
	project.Photo = filepath.Join(config.Current.HttpPathPhoto(), "project", fileName)
	return nil
}
