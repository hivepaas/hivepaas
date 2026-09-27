package mcp

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// The log tools read the log endpoints of apps, deployments and tasks with
// their own query parameters, and answer the lines as text a model reads:
// in time order, with their times, cut to what an answer holds. grep is the
// tools' own: the endpoints do not filter, so it is applied to what they
// answer.

const (
	// defaultLogTail is the lines a tool asks for when not told: fewer than
	// the endpoints' own 1000, which is more than a model reads well.
	defaultLogTail = 100
	// logBudget leaves room in MaxToolOutput for the rest of the answer.
	logBudget = MaxToolOutput - 1024
)

// descLogAnswer says what every log tool answers, and how grep works.
const descLogAnswer = "The answer is lines, how many lines the endpoint answered (answered) and, when older " +
	"lines were left out to keep the answer under 64 KB, how many (omitted). grep is the tool's own: it " +
	"asks the endpoint for the most lines it answers, 5000, within since and duration, keeps those that " +
	"match, and of them the newest tail; matched says how many matched."

// logParams are the query parameters the log endpoints share, and grep.
type logParams struct {
	Tail     int    `json:"tail,omitempty" jsonschema:"how many of the newest lines, 1-5000; 100 when not given"`
	Since    string `json:"since,omitempty" jsonschema:"only lines from this time on: a time in RFC 3339, such as 2026-09-26T08:00:00Z"`                                                               //nolint:lll
	Duration string `json:"duration,omitempty" jsonschema:"a length of time such as 30m or 2h: with since, only the lines in that span after since; alone, only the lines of that long ago until now"` //nolint:lll
	Grep     string `json:"grep,omitempty" jsonschema:"only lines containing this text, ignoring case; /expr/ for a regular expression"`                                                               //nolint:lll
}

func (p logParams) query() (*logQuery, error) {
	return newLogQuery(p.Tail, p.Since, p.Duration, p.Grep)
}

type appLogsInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string `json:"env" jsonschema:"the env's name, such as production; list_projects lists each project's envs"`
	App     string `json:"app" jsonschema:"the app's key, name or id; list_apps lists them"`
	TaskID  string `json:"taskId,omitempty" jsonschema:"only this container's lines: a task id (data[].id) from get_app_status"` //nolint:lll
	logParams
}

// logsAnswer is what a log tool answers.
type logsAnswer struct {
	Lines []string `json:"lines"`
	// Matched is how many of the lines answered matched grep, before the tail.
	Matched *int `json:"matched,omitempty"`
	// Answered is how many lines the endpoint answered.
	Answered int `json:"answered"`
	// Omitted is how many older lines were left out to keep the answer small.
	Omitted int `json:"omitted,omitempty"`
}

// logFrame is one line as the log endpoints answer it: tasklog's own frame.
type logFrame = tasklog.LogFrame

func getAppLogsTool() Tool {
	return readTool("get_app_logs", "Read an app's logs",
		"GET /projects/{project}/{env}/apps/{app}/logs. The newest lines the app's containers printed, as "+
			"Docker still holds them: the lines of every container together, oldest first, each with its time "+
			"and those written to stderr marked [stderr]. taskId reads one container. A container that is "+
			"gone takes its lines with it: search_app_logs searches what HivePaaS collected, when its logging "+
			"is on. "+descLogAnswer,
		func(ctx context.Context, call *Call, in appLogsInput) (logsAnswer, error) {
			q, err := in.query()
			if err != nil {
				return logsAnswer{}, err
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return logsAnswer{}, err
			}
			if taskID := strings.TrimSpace(in.TaskID); taskID != "" {
				q.extra = url.Values{"taskId": {taskID}}
			}
			return q.read(ctx, call, ref.path("/logs"))
		})
}

// logQuery is what a log tool was asked for.
type logQuery struct {
	tail     int
	since    time.Time
	duration string
	match    func(string) bool
	extra    url.Values
}

func newLogQuery(tail int, since, duration, grep string) (*logQuery, error) {
	if tail < 0 || tail > appdto.MaxAppLogsTail {
		return nil, &InputError{Message: fmt.Sprintf("tail is %d; it is 1 to %d", tail, appdto.MaxAppLogsTail)}
	}
	if tail == 0 {
		tail = defaultLogTail
	}
	q := &logQuery{tail: tail, duration: strings.TrimSpace(duration)}
	var err error
	if q.match, err = grepMatcher(grep); err != nil {
		return nil, err
	}
	if since = strings.TrimSpace(since); since != "" {
		if q.since, err = time.Parse(time.RFC3339, since); err != nil {
			return nil, &InputError{Message: fmt.Sprintf("since is %q; give a time in RFC 3339, such as "+
				"2026-09-26T08:00:00Z - or a length of time as duration, such as 2h", since)}
		}
	}
	return q, nil
}

// read asks a log endpoint for its lines, as the caller, and answers them.
func (q *logQuery) read(ctx context.Context, call *Call, path string) (logsAnswer, error) {
	// With grep, the endpoint is asked for all it answers: the lines that
	// match are then the newest that match, not those among the newest few.
	tail := q.tail
	if q.match != nil {
		tail = appdto.MaxAppLogsTail
	}
	query := url.Values{"tail": {strconv.Itoa(tail)}, "timestamps": {paramTrue}}
	if !q.since.IsZero() {
		query.Set("since", q.since.UTC().Format(time.RFC3339))
	}
	if q.duration != "" {
		query.Set("duration", q.duration)
	}
	for k, v := range q.extra {
		query[k] = v
	}
	var resp struct {
		Data struct {
			Logs []logFrame `json:"logs"`
		} `json:"data"`
	}
	if err := call.Get(ctx, path, query, &resp); err != nil {
		return logsAnswer{}, err
	}
	return makeLogsAnswer(resp.Data.Logs, q.tail, q.match), nil
}

// timeNow is time.Now, replaced in tests.
var timeNow = time.Now

// grepMatcher is the filter grep asks for, or nil for none.
func grepMatcher(grep string) (func(string) bool, error) {
	if grep == "" {
		return nil, nil
	}
	if len(grep) > 2 && strings.HasPrefix(grep, "/") && strings.HasSuffix(grep, "/") {
		re, err := regexp.Compile("(?i)" + grep[1:len(grep)-1])
		if err != nil {
			return nil, &InputError{Message: "grep is not a regular expression Go reads: " + err.Error()}
		}
		return re.MatchString, nil
	}
	lower := strings.ToLower(grep)
	return func(line string) bool { return strings.Contains(strings.ToLower(line), lower) }, nil
}

// makeLogsAnswer puts the lines in time order, filters them, keeps the newest
// tail of them, and of those what fits in an answer.
func makeLogsAnswer(frames []logFrame, tail int, match func(string) bool) logsAnswer {
	// An app's log is its containers' logs together, each in order but not
	// with one another: in time order, the newest lines are the last, which is
	// what the tail keeps.
	frames = slices.Clone(frames)
	slices.SortStableFunc(frames, func(a, b logFrame) int { return a.Ts.Compare(b.Ts) })
	out := logsAnswer{Answered: len(frames)}
	lines := make([]string, 0, len(frames))
	for _, f := range frames {
		text := strings.TrimRight(f.Data, "\r\n")
		if match != nil && !match(text) {
			continue
		}
		lines = append(lines, formatLogLine(f.Ts, string(f.Type), text))
	}
	if match != nil {
		matched := len(lines)
		out.Matched = &matched
	}
	if len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	out.Lines, out.Omitted = lastLines(lines, logBudget)
	return out
}

func formatLogLine(ts time.Time, typ, text string) string {
	var b strings.Builder
	if !ts.IsZero() {
		b.WriteString(ts.UTC().Format("2006-01-02T15:04:05.000Z") + " ")
	}
	if typ == string(tasklog.LogTypeErr) {
		b.WriteString("[stderr] ")
	}
	b.WriteString(text)
	return b.String()
}
