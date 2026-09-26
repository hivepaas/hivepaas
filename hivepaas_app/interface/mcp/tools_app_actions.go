package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ---- plan_restart_app ----

type restartPlan struct {
	App     string     `json:"app"`
	Project string     `json:"project"`
	Env     string     `json:"env"`
	Effect  string     `json:"effect"`
	Tasks   []taskItem `json:"tasksNow"`
}

func planRestartAppTool() Tool {
	return planTool("plan_restart_app", "Plan a restart",
		"Plans restarting an app: the swarm replaces its containers with new ones of the same image and "+
			"configuration, as the dashboard's Restart does. Answers its containers now. Nothing happens "+
			"until apply_plan.",
		NeedExecute, &applier{follow: "get_app_status shows the new containers start; get_app_logs what they print."},
		func(ctx context.Context, call *Call, in appInput) (restartPlan, *storedPlan, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return restartPlan{}, nil, err
			}
			var resp struct {
				Data []apiServiceTask `json:"data"`
			}
			if err = call.Get(ctx, ref.path("/service-tasks"), nil, &resp); err != nil {
				return restartPlan{}, nil, err
			}
			out := restartPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env,
				Effect: "every container of the app is replaced; the app may be unavailable while they start",
				Tasks:  makeAppStatus(ref, resp.Data).Tasks}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/restart"),
				Body:    json.RawMessage(`{}`),
				Summary: fmt.Sprintf("restart %s in %s/%s", ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}

// ---- plan_redeploy_app ----

type redeployInput struct {
	Project  string `json:"project" jsonschema:"the project's key, name or id"`
	Env      string `json:"env" jsonschema:"the env's name, such as prod"`
	App      string `json:"app" jsonschema:"the app's key, name or id"`
	ImageTag string `json:"imageTag,omitempty" jsonschema:"an image app: the tag to deploy; the current one when empty"`
	RepoRef  string `json:"repoRef,omitempty" jsonschema:"a repository app: the branch or tag to build"`
	NoCache  bool   `json:"noCache,omitempty" jsonschema:"a repository app: build without the build cache"`
}

// deploySource is where an app's image comes from, as its deployment settings
// say: what a redeploy plan shows, and what its apply checks has not moved.
type deploySource struct {
	Method  string `json:"method"`
	Image   string `json:"image,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Commit  string `json:"commit,omitempty"`
	NoCache bool   `json:"noCache,omitempty"`
}

type redeployPlan struct {
	App     string       `json:"app"`
	Project string       `json:"project"`
	Env     string       `json:"env"`
	Now     deploySource `json:"now"`
	After   deploySource `json:"after"`
	Effect  string       `json:"effect"`
	Issues  []string     `json:"issues,omitempty"`
}

const (
	methodImage = "image"
	methodRepo  = "repo"
)

func planRedeployAppTool() Tool {
	return planTool("plan_redeploy_app", "Plan a redeploy",
		"Plans deploying an app again: an image app pulls its image - or another tag of it - and a "+
			"repository app is built again from its branch, or another one. Answers the source now and "+
			"after. The source's other settings are configuration: plan_update_app_config changes them. "+
			"Nothing happens until apply_plan.",
		NeedExecute, &applier{check: checkDeploySource,
			follow: "get_app shows the new deployment and how it ends; get_task_logs its build."},
		func(ctx context.Context, call *Call, in redeployInput) (redeployPlan, *storedPlan, error) {
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return redeployPlan{}, nil, err
			}
			now, err := readDeploySource(ctx, call, ref.path(""))
			if err != nil {
				return redeployPlan{}, nil, err
			}
			return makeRedeployPlan(ref, now, &in)
		})
}

func makeRedeployPlan(ref *appRef, now deploySource, in *redeployInput) (redeployPlan, *storedPlan, error) {
	out := redeployPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Now: now, After: now,
		Effect: "a new deployment is made; the app's containers are replaced once it succeeds"}
	body := map[string]any{"activeMethod": now.Method}
	imageTag, repoRef := strings.TrimSpace(in.ImageTag), strings.TrimSpace(in.RepoRef)
	switch now.Method {
	case methodImage:
		if repoRef != "" || in.NoCache {
			out.Issues = append(out.Issues, "repoRef and noCache are for an app built from a repository; "+
				"this one runs an image")
		}
		if imageTag != "" {
			name, _, _ := strings.Cut(now.Image, ":")
			out.After.Image = name + ":" + imageTag
			body["imageSource"] = map[string]string{"imageTag": imageTag}
		}
	case methodRepo:
		if imageTag != "" {
			out.Issues = append(out.Issues, "imageTag is for an app that runs an image; this one is built "+
				"from a repository")
		}
		if repoRef != "" {
			out.After.Ref, out.After.Commit = repoRef, ""
			body["repoSource"] = map[string]string{"repoRef": repoRef}
		}
		out.After.NoCache = in.NoCache
		body["noCache"] = in.NoCache
	default:
		out.Issues = append(out.Issues, "the app has no source to deploy from yet; set one in the dashboard")
	}
	if len(out.Issues) > 0 {
		return out, nil, nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	seen, err := json.Marshal(redeployCheck{AppPath: ref.path(""), Seen: now})
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/deploy"), Body: raw, Check: seen,
		Summary: fmt.Sprintf("redeploy %s in %s/%s from %s", ref.AppKey, ref.ProjectKey, ref.Env,
			out.After.describe())}, nil
}

// describe is a source in a few words, for a summary.
func (s deploySource) describe() string {
	if s.Method == methodImage {
		return s.Image
	}
	return strings.TrimSpace(s.Repo + " " + s.Ref)
}

// apiDeploySettings is an app's deployment settings as the API answers them,
// as far as its source goes.
type apiDeploySettings struct {
	ActiveMethod string `json:"activeMethod"`
	ImageSource  *struct {
		Image string `json:"image"`
	} `json:"imageSource"`
	RepoSource *struct {
		RepoURL    string `json:"repoURL"`
		RepoRef    string `json:"repoRef"`
		CommitHash string `json:"commitHash"`
	} `json:"repoSource"`
}

// readDeploySource reads an app's source from its deployment settings: what the
// next deployment uses, which the last one may not have.
func readDeploySource(ctx context.Context, call *Call, appPath string) (deploySource, error) {
	var resp struct {
		Data apiDeploySettings `json:"data"`
	}
	if err := call.Get(ctx, appPath+"/deployment-settings", nil, &resp); err != nil {
		return deploySource{}, err
	}
	s := resp.Data
	out := deploySource{Method: s.ActiveMethod}
	if s.ImageSource != nil {
		out.Image = s.ImageSource.Image
	}
	if s.RepoSource != nil && s.ActiveMethod == methodRepo {
		out.Repo, out.Ref, out.Commit = withoutUserinfo(s.RepoSource.RepoURL), s.RepoSource.RepoRef,
			s.RepoSource.CommitHash
	}
	if s.ActiveMethod != methodImage {
		out.Image = ""
	}
	return out, nil
}

// redeployCheck is what a redeploy plan keeps to check: the source it saw, and
// the app it is of.
type redeployCheck struct {
	AppPath string       `json:"appPath"`
	Seen    deploySource `json:"seen"`
}

// checkDeploySource refuses a redeploy whose app's source changed since the
// plan: the person agreed to deploy what they were shown.
func checkDeploySource(ctx context.Context, call *Call, raw json.RawMessage) error {
	var check redeployCheck
	if err := json.Unmarshal(raw, &check); err != nil {
		return fmt.Errorf("mcp: reading a plan's check: %w", err)
	}
	now, err := readDeploySource(ctx, call, check.AppPath)
	if err != nil {
		return err
	}
	if now != check.Seen {
		return planMoved("the app's source")
	}
	return nil
}
