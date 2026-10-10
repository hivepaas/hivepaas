// Package taskdatafileload feeds an app's data file to a command run in the
// app, on its stdin: a dump a job saved, loaded back into the app's database.
//
// It is a task because a load takes as long as the file is large, and what it
// did has to be told in a log a person can read afterwards.
package taskdatafileload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/rediscache"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/filecodec"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Executor struct {
	db          *database.DB
	redisClient rediscache.Client

	fileRepo    repository.FileRepo
	taskLogRepo repository.TaskLogRepo

	appService          appservice.Service
	fileService         fileservice.Service
	schedJobExecService schedjobexecservice.Service
	settingService      settingservice.Service
}

func NewExecutor(
	taskQueue queue.TaskQueue,
	db *database.DB,
	redisClient rediscache.Client,
	fileRepo repository.FileRepo,
	taskLogRepo repository.TaskLogRepo,
	appService appservice.Service,
	fileService fileservice.Service,
	schedJobExecService schedjobexecservice.Service,
	settingService settingservice.Service,
) *Executor {
	e := &Executor{
		db:                  db,
		redisClient:         redisClient,
		fileRepo:            fileRepo,
		taskLogRepo:         taskLogRepo,
		appService:          appService,
		fileService:         fileService,
		schedJobExecService: schedJobExecService,
		settingService:      settingService,
	}
	taskQueue.RegisterExecutor(base.TaskTypeDataFileLoad, e.execute)
	return e
}

func (e *Executor) execute(
	ctx context.Context,
	db database.Tx,
	task *queue.TaskExecData,
) (err error) {
	// A load half done is not done again by itself: what it left is for a person
	// to look at.
	task.TaskNonRetryable = true
	task.LogStore = tasklog.NewRemoteStore(fmt.Sprintf("task:%s:log", task.Task.ID), e.redisClient)
	task.LogStore.SetOnFlush(tasklog.DefaultMaxSize, func(ctx context.Context, frames []*tasklog.LogFrame) error {
		return e.saveLogFrames(ctx, e.db, task.Task, frames)
	})
	task.OnPostTx(func() { _ = e.saveLogs(context.Background(), e.db, task) }) //nolint:contextcheck
	defer func() {
		if err != nil {
			_ = task.LogStore.Add(ctx, tasklog.NewErrFrame("Load failed: "+err.Error()+"\n", tasklog.TsNow))
		}
		_ = task.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Load finished in %s\n",
			timeutil.NowUTC().Sub(task.Task.StartedAt).Truncate(time.Millisecond)), tasklog.TsNow))
		_ = e.saveLogs(ctx, db, task)
	}()
	defer safego.RecoverTo(&err)

	args, err := task.Task.ArgsAsDataFileLoad()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if args == nil || args.Command == nil {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("data file load task has no command")
	}

	// The file may have been deleted since the load was asked for.
	file, err := e.fileRepo.GetByID(ctx, db, task.Task.TargetID,
		bunex.SelectWhere("file.status = ?", base.FileStatusActive),
		bunex.SelectWhere("file.object_id = ?", args.AppID),
	)
	if err != nil {
		return hperrors.Wrap(err).WithExtraDetail("the file to load is gone")
	}
	if file.Deleted {
		return hperrors.NewNotFound("File").WithExtraDetail("the file to load is gone")
	}
	app, err := e.appService.LoadApp(ctx, db, args.ProjectID, args.AppID, true, true,
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// The command's script, and what its env vars reference.
	if task.RefObjects == nil {
		task.RefObjects = entity.NewRefObjects()
	}
	err = e.settingService.LoadRefObjectsByIDs(ctx, db, &task.RefObjects, app.GetObjectScope(), true,
		args.Command.GetRefObjectIDs())
	if err != nil {
		return hperrors.Wrap(err)
	}
	passphrase, err := args.Passphrase.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}

	content, err := e.fileService.Open(ctx, db, file)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = content.Close() }()
	decoded, err := filecodec.NewReader(content, file.Name, passphrase)
	if err != nil {
		return hperrors.Wrap(err).WithExtraDetail("the file cannot be read")
	}

	_ = task.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Loading %s (%s) into %s\n",
		file.Name, unit.DataSize(file.Size).HR(), app.Name), tasklog.TsNow))
	stdin := &readingReader{r: decoded}
	_, err = e.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: task, App: app, Command: args.Command, Stdin: stdin,
	})
	// The command read a part of the file only: whatever it says, the load failed.
	read, readErr := stdin.read()
	if readErr != nil {
		err = errors.Join(err, hperrors.Wrap(readErr).WithExtraDetail("the file could not be read whole"))
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = task.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Loaded %s into the command\n",
		unit.DataSize(read).HR()), tasklog.TsNow))
	return nil
}

// readingReader counts what is read of the file, and keeps the error that
// ended the reading before its end. The command's stdin is copied from it on
// a goroutine of its own, which may outlast the command.
type readingReader struct {
	r io.Reader

	mu  sync.Mutex
	n   int64
	err error
}

func (r *readingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.n += int64(n)
	if err != nil && !errors.Is(err, io.EOF) && r.err == nil {
		r.err = err
	}
	return n, err //nolint:wrapcheck // a reader's errors pass through as they are
}

// read is how much was read, and the error that ended it early, if one did.
func (r *readingReader) read() (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n, r.err
}

func (e *Executor) saveLogs(ctx context.Context, db database.IDB, task *queue.TaskExecData) error {
	frames, err := task.LogStore.GetData(ctx, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = task.LogStore.Reset() //nolint:contextcheck // the store resets itself
	return e.saveLogFrames(ctx, db, task.Task, frames)
}

func (e *Executor) saveLogFrames(
	ctx context.Context,
	db database.IDB,
	task *entity.Task,
	frames []*tasklog.LogFrame,
) error {
	for _, chunk := range gofn.Chunk(frames, 10000) { //nolint:mnd
		taskLogs := make([]*entity.TaskLog, 0, len(chunk))
		for _, frame := range chunk {
			taskLogs = append(taskLogs, &entity.TaskLog{
				TaskID: task.ID, TargetID: task.TargetID, Type: frame.Type, Data: frame.Data, Ts: frame.Ts,
			})
		}
		if err := e.taskLogRepo.InsertMulti(ctx, db, taskLogs); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
