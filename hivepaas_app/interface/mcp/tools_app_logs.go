package mcp

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultLogTail and maxLogTail bound the lines get_app_logs answers.
	defaultLogTail = 100
	maxLogTail     = 500
	// grepScan is how many lines a grep reads through: the most the log
	// endpoint answers at once.
	grepScan = 5000
	// logBudget leaves room in MaxToolOutput for the rest of the answer.
	logBudget = MaxToolOutput - 1024
)

type appLogsInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	Tail    int    `json:"tail,omitempty" jsonschema:"how many of the newest lines to answer, 1-500; 100 when not given"`
	Since   string `json:"since,omitempty" jsonschema:"only lines since then: an RFC 3339 time, or a duration ago like 2h"`
	Grep    string `json:"grep,omitempty" jsonschema:"only lines containing this, ignoring case; /expr/ for a regexp"`
}

type appLogs struct {
	App   string   `json:"app"`
	Lines []string `json:"lines"`
	// Matched is how many of the lines read matched grep, before the tail.
	Matched *int `json:"matched,omitempty"`
	// Scanned is how many lines grep read through.
	Scanned int `json:"scanned,omitempty"`
	// Omitted is how many older lines were left out to keep the answer small.
	Omitted int `json:"omitted,omitempty"`
}

// logFrame is one line as the log endpoint answers it.
type logFrame struct {
	Type string    `json:"type"`
	Data string    `json:"data"`
	Ts   time.Time `json:"ts"`
}

func getAppLogsTool() Tool {
	return readTool("get_app_logs", "Read an app's logs",
		"Reads the newest lines an app's containers printed, with their times; lines written to "+
			"stderr are marked. With grep, it reads up to the last 5000 lines and answers the newest "+
			"of those that match, so an error from an hour ago is found under a busy log.",
		func(ctx context.Context, call *Call, in appLogsInput) (appLogs, error) {
			if err := in.check(); err != nil {
				return appLogs{}, err
			}
			match, err := grepMatcher(in.Grep)
			if err != nil {
				return appLogs{}, err
			}
			since, err := parseSince(in.Since, timeNow())
			if err != nil {
				return appLogs{}, err
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return appLogs{}, err
			}
			frames, err := readLogs(ctx, call, ref.path("/logs"), in.tail(match != nil), since)
			if err != nil {
				return appLogs{}, err
			}
			return makeAppLogs(ref.AppKey, frames, in.tail(false), match), nil
		})
}

func (in *appLogsInput) check() error {
	if in.Tail < 0 || in.Tail > maxLogTail {
		return &InputError{Message: fmt.Sprintf("tail is %d; it is 1 to %d", in.Tail, maxLogTail)}
	}
	return nil
}

// tail is how many lines to ask the endpoint for: all it answers when they are
// to be filtered, since the lines that match may be anywhere in them.
func (in *appLogsInput) tail(grep bool) int {
	switch {
	case grep:
		return grepScan
	case in.Tail == 0:
		return defaultLogTail
	}
	return in.Tail
}

// timeNow is time.Now, replaced in tests.
var timeNow = time.Now

// parseSince reads a time in RFC 3339, or a duration back from now.
func parseSince(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return now.Add(-d), nil
	}
	return time.Time{}, &InputError{Message: fmt.Sprintf(
		"since is %q; give a time in RFC 3339 such as 2026-09-26T08:00:00Z, or a duration such as 2h", s)}
}

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

// readLogs asks a log endpoint for its newest lines, as the caller.
func readLogs(ctx context.Context, call *Call, path string, tail int, since time.Time) ([]logFrame, error) {
	query := url.Values{"tail": {strconv.Itoa(tail)}, "timestamps": {paramTrue}}
	if !since.IsZero() {
		query.Set("since", since.UTC().Format(time.RFC3339))
	}
	var resp struct {
		Data struct {
			Logs []logFrame `json:"logs"`
		} `json:"data"`
	}
	if err := call.Get(ctx, path, query, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Logs, nil
}

// makeAppLogs filters the frames, keeps the newest tail of them, and keeps of
// those what fits in an answer.
func makeAppLogs(app string, frames []logFrame, tail int, match func(string) bool) appLogs {
	out := appLogs{App: app}
	lines := make([]string, 0, len(frames))
	for _, f := range frames {
		text := strings.TrimRight(f.Data, "\r\n")
		if match != nil && !match(text) {
			continue
		}
		lines = append(lines, formatLogLine(f.Ts, f.Type, text))
	}
	if match != nil {
		matched := len(lines)
		out.Matched, out.Scanned = &matched, len(frames)
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
	if typ == "err" {
		b.WriteString("[stderr] ")
	}
	b.WriteString(text)
	return b.String()
}
