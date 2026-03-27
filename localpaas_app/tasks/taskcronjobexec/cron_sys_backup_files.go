package taskcronjobexec

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jsonl"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	fileModelUserPhoto    = "files/user-photo"
	fileModelProjectPhoto = "files/project-photo"
)

const (
	sysBackupFilePageDataSize = 10 * 1024 * 1024 // KB
)

var (
	sysBackupFileModels = []*sysBackupFileModel{
		{
			Type: fileModelUserPhoto,
			DirPath: func() string {
				return config.Current.DataPathUserPhoto()
			},
		},
		{
			Type: fileModelProjectPhoto,
			DirPath: func() string {
				return config.Current.DataPathProjectPhoto()
			},
		},
	}
)

type sysBackupFileModel struct {
	Type         string
	PageDataSize int
	DirPath      func() string
}

type sysBackupFileEntry struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

//nolint:gocognit
func (e *Executor) sysBackupFiles(
	ctx context.Context,
	db database.IDB,
	jsonlW *jsonl.Writer,
	data *sysBackupTaskData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewWarnFrame("Start backing up static files...", applog.TsNow))

	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewWarnFrame("Files backup finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Files backup finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	allUsers, _, err := e.userRepo.List(ctx, db, nil, bunex.SelectColumns("id"))
	if err != nil {
		return apperrors.Wrap(err)
	}
	allUserIDs := entityutil.SliceToIDMap(allUsers)

	allProjects, _, err := e.projectRepo.List(ctx, db, nil, bunex.SelectColumns("id"))
	if err != nil {
		return apperrors.Wrap(err)
	}
	allProjectIDs := entityutil.SliceToIDMap(allProjects)

	for _, model := range sysBackupFileModels {
		dirPath := model.DirPath()
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return apperrors.Wrap(err)
		}

		maxPageDataSize := gofn.Coalesce(model.PageDataSize, sysBackupFilePageDataSize)
		pageDataSize := 0
		savingData := make([]*sysBackupFileEntry, 0, 20) //nolint:mnd
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fileName := entry.Name()
			switch model.Type {
			case fileModelUserPhoto:
				userID, _, _ := strings.Cut(fileName, ".")
				if _, exists := allUserIDs[userID]; !exists {
					continue
				}
			case fileModelProjectPhoto:
				projectID, _, _ := strings.Cut(fileName, ".")
				if _, exists := allProjectIDs[projectID]; !exists {
					continue
				}
			default:
				// Allow
			}

			fileData, err := os.ReadFile(filepath.Join(dirPath, fileName))
			if err != nil {
				return apperrors.Wrap(err)
			}
			savingData = append(savingData, &sysBackupFileEntry{
				Name: entry.Name(),
				Data: fileData,
			})
			pageDataSize += len(fileData)
			if pageDataSize >= maxPageDataSize {
				err = jsonlW.WriteChunk(jsonl.NewChunk(model.Type, savingData))
				if err != nil {
					return apperrors.Wrap(err)
				}
				savingData = savingData[:0]
				pageDataSize = 0
			}
		}
		// Last page of data
		if len(savingData) > 0 {
			err = jsonlW.WriteChunk(jsonl.NewChunk(model.Type, savingData))
			if err != nil {
				return apperrors.Wrap(err)
			}
		}
	}

	return nil
}
