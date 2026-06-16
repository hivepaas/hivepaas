package entity

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentTaskVersion = 1
)

var (
	TaskUpsertingConflictCols = []string{"id"}
	TaskUpsertingUpdateCols   = []string{"target_id", "type", "status", "config",
		"args", "runs", "output", "version", "update_ver", "run_at", "retry_at",
		"started_at", "ended_at", "updated_at", "deleted_at"}
)

type Task struct {
	ID        string          `bun:",pk" json:"id"`
	TargetID  string          `bun:",nullzero" json:"targetId"`
	Type      base.TaskType   `json:"type"`
	Status    base.TaskStatus `json:"status"`
	Config    TaskConfig      `bun:",nullzero" json:"config"`
	Args      string          `bun:",nullzero" json:"args"`
	Runs      string          `bun:",nullzero" json:"runs,omitempty"`
	Output    string          `bun:",nullzero" json:"output,omitempty"`
	Version   int             `json:"version"`
	UpdateVer int             `json:"updateVer"`

	RunAt     time.Time `bun:",nullzero" json:"runAt"`
	RetryAt   time.Time `bun:",nullzero" json:"retryAt"`
	StartedAt time.Time `bun:",nullzero" json:"startedAt"`
	EndedAt   time.Time `bun:",nullzero" json:"endedAt"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
	DeletedAt time.Time `bun:",soft_delete,nullzero" json:"deletedAt,omitzero"`

	TargetJob        *Setting    `bun:"rel:belongs-to,join:target_id=id" json:"targetJob,omitempty"`
	TargetDeployment *Deployment `bun:"rel:belongs-to,join:target_id=id" json:"targetDeployment,omitempty"`

	// NOTE: temporary fields
	parsedArgs   any
	parsedOutput any
}

type TaskConfig struct {
	Priority        base.TaskPriority `json:"priority"`
	MaxRetry        int               `json:"maxRetry,omitempty"`
	Retry           int               `json:"retry,omitempty"`
	RetryDelay      timeutil.Duration `json:"retryDelay,omitempty"`
	Timeout         timeutil.Duration `json:"timeout,omitempty"`
	ControlDisabled bool              `json:"controlDisabled,omitempty"`
}

// GetID implements IDEntity interface
func (t *Task) GetID() string {
	return t.ID
}

func (t *Task) IsNotStarted() bool {
	return t.Status == base.TaskStatusNotStarted
}

func (t *Task) IsInProgress() bool {
	return t.Status == base.TaskStatusInProgress
}

func (t *Task) IsDone() bool {
	return t.Status == base.TaskStatusDone
}

func (t *Task) IsFailedCompletely() bool {
	return t.Status == base.TaskStatusFailed && !t.CanRetry()
}

func (t *Task) IsCanceled() bool {
	return t.Status == base.TaskStatusCanceled
}

func (t *Task) CanCancel() bool {
	return t.IsNotStarted() || t.IsInProgress()
}

func (t *Task) CanRetry() bool {
	return t.Status == base.TaskStatusFailed && t.Config.MaxRetry > t.Config.Retry
}

func (t *Task) CreateTimeoutCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	if t.Config.Timeout > 0 {
		return context.WithTimeout(ctx, t.Config.Timeout.ToDuration())
	}
	return ctx, func() {}
}

func (t *Task) GetDuration() time.Duration {
	if t.EndedAt.IsZero() || t.EndedAt.Before(t.StartedAt) {
		return time.Since(t.StartedAt)
	}
	return t.EndedAt.Sub(t.StartedAt)
}

func (t *Task) ShouldRunAt() (runAt time.Time) {
	runAt = t.RunAt
	if t.Status == base.TaskStatusFailed {
		runAt = t.RetryAt
	}
	if !runAt.IsZero() && runAt.Before(timeutil.NowUTC()) {
		runAt = time.Time{}
	}
	return runAt
}

func (t *Task) GetRuns() ([]*TaskRun, error) {
	if t.Runs == "" {
		return nil, nil
	}
	runs := []*TaskRun{}
	err := json.Unmarshal(reflectutil.UnsafeStrToBytes(t.Runs), &runs)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return runs, nil
}

func (t *Task) AddRun(run *TaskRun) error {
	runs, err := t.GetRuns()
	if err != nil {
		return apperrors.Wrap(err)
	}

	runs = append(runs, run)
	runBytes, err := json.Marshal(runs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	t.Runs = reflectutil.UnsafeBytesToStr(runBytes)
	return nil
}

type TaskRun struct {
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	Error     string    `json:"error,omitempty"`
}

func (t *Task) parseArgs(structPtr any) error {
	if t == nil || len(t.Args) == 0 {
		return nil
	}
	err := json.Unmarshal(reflectutil.UnsafeStrToBytes(t.Args), structPtr)
	if err != nil {
		return apperrors.Wrap(err)
	}
	t.parsedArgs = structPtr
	return nil
}

func (t *Task) SetArgs(args any) error {
	b, err := json.Marshal(args)
	if err != nil {
		return apperrors.Wrap(err)
	}
	t.Args = reflectutil.UnsafeBytesToStr(b)
	t.parsedArgs = args
	return nil
}

func (t *Task) MustSetArgs(args any) {
	gofn.Must1(t.SetArgs(args))
}

func parseTaskArgsAs[T any](t *Task, newFn func() T) (res T, error error) {
	if t.parsedArgs != nil {
		res, ok := t.parsedArgs.(T)
		if !ok {
			return res, apperrors.NewMismatch("Task args type", reflect.TypeFor[T]())
		}
		return res, nil
	}
	if t.Args != "" {
		res = newFn()
		if err := t.parseArgs(res); err != nil {
			return res, apperrors.Wrap(err)
		}
	}
	return res, nil
}

func (t *Task) parseOutput(structPtr any) error {
	if t == nil || t.Output == "" {
		return nil
	}
	err := json.Unmarshal(reflectutil.UnsafeStrToBytes(t.Output), structPtr)
	if err != nil {
		return apperrors.Wrap(err)
	}
	t.parsedOutput = structPtr
	return nil
}

func (t *Task) SetOutput(output any) error {
	b, err := json.Marshal(output)
	if err != nil {
		return apperrors.Wrap(err)
	}
	t.Output = reflectutil.UnsafeBytesToStr(b)
	t.parsedOutput = output
	return nil
}

func (t *Task) MustSetOutput(output any) {
	gofn.Must1(t.SetOutput(output))
}

func parseTaskOutputAs[T any](t *Task, newFn func() T) (res T, error error) {
	if t.parsedOutput != nil {
		res, ok := t.parsedOutput.(T)
		if !ok {
			return res, apperrors.NewMismatch("Task output type", reflect.TypeFor[T]())
		}
		return res, nil
	}
	if t.Output != "" {
		res = newFn()
		if err := t.parseOutput(res); err != nil {
			return res, apperrors.Wrap(err)
		}
	}
	return res, nil
}
