package mcp

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/homeuc/homedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc/taskdto"
)

// The read tools that are one GET endpoint each. What each answers is the
// endpoint's own response - {meta, data} - and what each takes is the
// endpoint's own query parameters, with names for what its path needs ids for.

// Descriptions of the parameters many endpoints share.
const (
	descSearchNameNote = "text in the name or the note, ignoring case; * matches anything"
	descSearchName     = "text in the name, ignoring case; * matches anything"
	descTemplate       = "the template's name, from search_templates"
	descPageOffset     = "how many to skip, for the next page; meta.page.total says how many there are in all"
	descPageLimit      = "how many to answer, 1-10000; 50 when not given"
	descSort           = "fields to order by, comma-separated, each ascending or with - before it descending, " +
		"such as -createdAt"
)

func pagingParams(extra map[string]string) map[string]string {
	out := map[string]string{paramPageOffset: descPageOffset, paramPageLimit: descPageLimit, paramSort: descSort}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func endpointTools() []Tool {
	endpoints := getEndpoints()
	out := make([]Tool, 0, len(endpoints))
	for _, e := range endpoints {
		out = append(out, e.tool())
	}
	return out
}

// settingStatuses are the states of an object kept as a setting, such as a
// node or a scheduled job.
var settingStatuses = statusValues(gofn.Drop(base.AllSettingStatuses, base.SettingStatusMissing))

var projectStatuses = statusValues(gofn.Drop(base.AllProjectStatuses, base.ProjectStatusMissing))

func getEndpoints() []getEndpoint { //nolint:funlen // a table
	return append([]getEndpoint{
		{
			name: "list_projects", title: "List projects",
			description: "GET /projects. Lists the projects the API key's user can see, each with its envs " +
				"(data[].envs[].name is what the other tools take as env). Start here to find the project and " +
				"env an app is in.",
			paths: map[under]string{underNothing: "/projects"},
			query: &projectdto.ListProjectReq{},
			params: pagingParams(map[string]string{paramSearch: descSearchNameNote,
				paramStatus: "only projects in these states: " + projectStatuses}),
			answer: func() any { return &projectdto.ListProjectResp{} },
		},
		{
			name: "list_apps", title: "List an env's apps",
			description: "GET /projects/{project}/{env}/apps. Lists the apps of one env of a project: each " +
				"app's key and name, status, category (a function, a database...), engine (the database or " +
				"cache it is, if any) and links. An app made by another - a template's component or " +
				"dependency, a preview - is listed inside it, under childApps and logicalChildApps, when " +
				"getChildApps is true.",
			paths: map[under]string{underEnv: "/apps"},
			query: &appdto.ListAppReq{},
			params: pagingParams(map[string]string{
				paramSearch:    descSearchNameNote,
				paramStatus:    "only apps in these states: " + statusValues(base.AllAppStatuses),
				"category":     "only apps of these categories, " + statusValues(base.AllAppCategories) + "; no kind is webapp",
				"parentId":     "only the apps made by the app of this id, such as its previews",
				paramGetStats:  "true to answer each app's running, desired and completed containers, under stats",
				"getChildApps": "true to answer the apps each app made, inside it",
			}),
			answer: func() any { return &appdto.ListAppResp{} },
		},
		{
			name: "get_app", title: "Get an app",
			description: "GET /projects/{project}/{env}/apps/{app}. One app: its key, name, status, engine, " +
				"note, tags, links, the app that made it (parentApp) and the apps it made (childApps, " +
				"logicalChildApps). How its containers run is get_app_status; its settings, get_app_settings; " +
				"its deployments, list_app_deployments.",
			paths: map[under]string{underApp: ""},
			query: &appdto.GetAppReq{},
			params: map[string]string{
				paramGetStats: "true to answer its running, desired and completed containers, under stats",
			},
			answer: func() any { return &appdto.GetAppResp{} },
		},
		{
			name: "get_app_status", title: "Get an app's containers",
			description: "GET /projects/{project}/{env}/apps/{app}/service-tasks. The app's swarm tasks - its " +
				"containers, past and present - those meant to run first, then by state, slot, and newest " +
				"first: the node each is on, the state it is in and the state it should be in, and the error " +
				"that stopped it. This is where a container that never " +
				"starts - an image that cannot be pulled, a port taken, a constraint no node meets - says why.",
			paths: map[under]string{underApp: "/service-tasks"},
			query: &appsettingsdto.GetAppServiceTasksReq{},
			params: map[string]string{"state": "only tasks in these swarm states, such as running, failed, " +
				"rejected, shutdown, complete; running, complete, shutdown and failed when not given"},
			answer: func() any { return &appsettingsdto.GetAppServiceTasksResp{} },
		},
		{
			name: "list_app_deployments", title: "List an app's deployments",
			description: "GET /projects/{project}/{env}/apps/{app}/deployments. The app's deployments, newest " +
				"first: status, what started it (trigger), the source it deployed (settings), when it started " +
				"and ended, and its output - the commit it built, the error that stopped it. " +
				"get_app_deployment_logs reads one's log.",
			paths: map[under]string{underApp: "/deployments"},
			query: &appdeploymentdto.ListDeploymentReq{},
			omit:  []string{paramSearch}, // taken, but the endpoint does not search yet
			params: pagingParams(map[string]string{
				paramStatus: "only deployments in these states: " + statusValues(base.AllDeploymentStatuses),
			}),
			answer: func() any { return &appdeploymentdto.ListDeploymentResp{} },
		},
		{
			name: "get_app_deployment", title: "Get a deployment",
			description: "GET /projects/{project}/{env}/apps/{app}/deployments/{deployment}. One deployment, as " +
				"list_app_deployments answers each. A deployment still running shows how far it got; ask again " +
				"to follow it, and get_app_deployment_logs for its log.",
			paths: map[under]string{underApp: "/deployments/{deployment}"},
			item: &pathItem{arg: "deployment", description: "the deployment's id, from list_app_deployments or " +
				"the redeploy that made it"},
			answer: func() any { return &appdeploymentdto.GetDeploymentResp{} },
		},
		{
			name: "search_app_logs", title: "Search an app's stored logs",
			description: "GET /projects/{project}/{env}/apps/{app}/logs/history. Searches the logs HivePaaS has " +
				"collected for the app - of containers that are gone as well as those running - over a time " +
				"range: an hour up to now when not given. data.logs are the newest matching lines within the " +
				"range, up to limit, oldest first; when there were more, data.truncated is true and " +
				"data.nextEnd is the end that reads the page before. The endpoint refuses it as unavailable " +
				"when the app's logging feature is off (get_app_settings, kind feature). get_app_logs reads " +
				"what Docker holds for the containers now instead.",
			paths: map[under]string{underApp: "/logs/history"},
			query: &appdto.GetAppLogHistoryReq{},
			params: map[string]string{
				"start": "the range's start, RFC 3339; an hour before end when not given",
				"end":   "the range's end, RFC 3339; now when not given. To page back, the answer's nextEnd",
				"limit": "how many lines, 1-5000; 500 when not given",
				paramSearch: "text to find in a line; matched from the start of a word, which the index answers " +
					"fast. With regex true it is a regular expression, read line by line: slower",
				"regex":     "true to read search as a regular expression",
				"matchCase": "true to compare search case-sensitively",
				"levels":    "only lines of these levels: trace, debug, info, warn, warning, error, fatal, panic",
				"streams":   "only lines of these streams: stdout, stderr",
			},
			answer: func() any { return &appdto.GetAppLogHistoryResp{} },
		},
		{
			name: "get_function_metrics", title: "Get a function's calls",
			description: "GET /projects/{project}/{env}/apps/{app}/function-metrics. A function's calls over a " +
				"range ending now, counted from the line its runtime logs for every call: how many, how many " +
				"failed (an outcome other than ok), how many its handler answered 5xx, and its duration's p50, " +
				"p95 and p99 in milliseconds - close, not exact - in totals, by outcome, and as series, a point " +
				"per step, oldest first. available is false, with a reason, when its logs cannot be read: the " +
				"app's logging feature off, or the logging stack not running. Only a function has them.",
			paths:  map[under]string{underApp: "/function-metrics"},
			query:  &appdto.GetFunctionMetricsReq{},
			params: map[string]string{"range": "1h, 6h, 24h or 7d; 24h when not given"},
			answer: func() any { return &appdto.GetFunctionMetricsResp{} },
		},
		{
			name: "list_attention", title: "What needs attention",
			description: "GET /home/attention. What the dashboard's Home page says needs attention and the API " +
				"key's user may see: apps whose containers do not all run or keep restarting, nodes that are " +
				"down, memory promised past what a node has - each with its kind, severity, what it is about, " +
				"and its last error. Start here when asked what is wrong.",
			paths:  map[under]string{underNothing: "/home/attention"},
			answer: func() any { return &homedto.GetHomeAttentionResp{} },
		},
		{
			name: "list_tasks", title: "List background tasks",
			description: "GET /system/tasks, /projects/{project}/{env}/tasks or " +
				"/projects/{project}/{env}/apps/{app}/tasks: the newest background tasks - deployments, " +
				"scheduled job runs, backups, certificate renewals, cleanups - with their state and the error " +
				"that ended a failed one. Given no project, every task the API key's user may see; given " +
				"project and env, that env's; given an app as well, that app's. get_task_logs reads what one " +
				"printed.",
			paths: map[under]string{underNothing: "/system/tasks", underEnv: "/tasks", underApp: "/tasks"},
			query: &taskdto.ListTaskReq{},
			params: pagingParams(map[string]string{
				paramSearch:   "text in the task's type, ignoring case; * matches anything",
				"type":        "only tasks of these types: " + statusValues(base.AllTaskTypes),
				paramStatus:   "only tasks in these states: " + statusValues(base.AllTaskStatuses),
				"targetId":    "only the tasks run for these objects, by id: a scheduled job's id for its runs",
				paramFromDate: "only tasks created on or after this date, YYYY-MM-DD",
				paramToDate:   "only tasks created on or before this date, YYYY-MM-DD",
				"scopeOnly": "true for the tasks of the scope itself only, not those of what is in it - an " +
					"env's own, not its apps'",
				"projectId":    "with no project given: only the tasks of the project of this id",
				"projectEnvId": "with no project given: only the tasks of this env, by its id",
				"appId":        "with no project given: only the tasks of the app of this id",
			}),
			answer: func() any { return &taskdto.ListTaskResp{} },
		},
		{
			name: "get_task", title: "Get a background task",
			description: "GET /system/tasks/{task}, or the same under an env or app when they are given: one " +
				"task - its type, state, what it ran for (targetJob), when it ran and its last error.",
			paths: map[under]string{underNothing: "/system/tasks/{task}", underEnv: "/tasks/{task}",
				underApp: "/tasks/{task}"},
			item:   &pathItem{arg: "task", description: "the task's id, from list_tasks"},
			answer: func() any { return &taskdto.GetTaskResp{} },
		},
		{
			name: "list_nodes", title: "List cluster nodes",
			description: "GET /cluster/nodes. The swarm's nodes: role, whether it leads, state, availability, " +
				"address, platform, CPUs and memory, Docker's version, and labels - what placement " +
				"constraints match against.",
			paths: map[under]string{underNothing: "/cluster/nodes"},
			query: &nodedto.ListNodeReq{},
			params: pagingParams(map[string]string{
				paramSearch: descSearchName,
				paramStatus: "only nodes whose record is in these states: " + settingStatuses,
				argKind:     "only nodes of these swarm roles: manager, worker",
			}),
			answer: func() any { return &nodedto.ListNodeResp{} },
		},
		{
			name: "list_volumes", title: "List volumes",
			description: "GET /projects/{project}[/{env}]/cluster-volumes, or /cluster/volumes when no project is " +
				"given: the swarm volumes an env's apps may use - each with its id, name, driver, the node it is " +
				"pinned to, and how many containers use it (refCount). A template parameter of type volume, such " +
				"as a database's dataVolume, takes one's id.",
			paths: map[under]string{underNothing: "/cluster/volumes", underProject: "/cluster-volumes",
				underEnv: "/cluster-volumes"},
			query: &volumedto.ListVolumeReq{},
			params: pagingParams(map[string]string{
				paramSearch: descSearchName,
				paramStatus: "only volumes whose record is in these states: " + settingStatuses,
				argKind:     "only volumes of these drivers, such as local",
			}),
			answer: func() any { return &volumedto.ListVolumeResp{} },
		},
		{
			name: "get_template_catalog", title: "The app store's categories and tags",
			description: "GET /app-templates/catalog. The categories and tags the app store's templates are " +
				"filed under - what search_templates takes as category and tag - and where the store is read " +
				"from (source, revision).",
			paths:  map[under]string{underNothing: "/app-templates/catalog"},
			answer: func() any { return &apptemplatedto.GetAppTemplateCatalogResp{} },
		},
		{
			name: "search_templates", title: "Search the app store",
			description: "GET /app-templates. The app store's templates, in name order: what each installs, " +
				"its versions and variants, the apps it brings along (dependencies) or is made of " +
				"(components), and whether it needs the Docker API or extra capabilities. A template that is " +
				"not compatible needs a newer HivePaaS. get_template shows one in full.",
			paths: map[under]string{underNothing: "/app-templates"},
			query: &apptemplatedto.ListAppTemplatesReq{},
			omit:  []string{paramSort}, // the order is by name and cannot be changed
			params: map[string]string{
				paramSearch: "text in a template's name, title, tagline, tags or aliases, ignoring case",
				"category": "only templates in any of these categories, such as databases or databases/sql; a " +
					"parent matches every category under it. get_template_catalog lists them",
				"tag":           "only templates carrying any of these tags; get_template_catalog lists them",
				paramPageOffset: descPageOffset, paramPageLimit: descPageLimit,
			},
			answer: func() any { return &apptemplatedto.ListAppTemplatesResp{} },
		},
		{
			name: "get_template", title: "Get a template",
			description: "GET /app-templates/{template}. One template in full: its description, the parameters " +
				"an install asks for - a secret one left empty is generated - its versions and variants, its " +
				"dependencies and components, and what it is granted: Docker API access, capabilities, " +
				"published ports.",
			paths:  map[under]string{underNothing: "/app-templates/{template}"},
			item:   &pathItem{arg: argTemplate, description: descTemplate},
			answer: func() any { return &apptemplatedto.GetAppTemplateResp{} },
		},
		{
			name: "list_template_image_tags", title: "List a template's image tags",
			description: "GET /app-templates/{template}/image-tags. Other builds of the image a version of a " +
				"template pins, as the registry publishes them - an install may take one as imageTag instead of " +
				"the pinned tag (currentTag). Only pinned tags of the same base (alpine, bookworm, ...) are " +
				"listed, newest first. Each tag's class is same-line - another release of the version line the " +
				"template declares, whose configuration applies but which HivePaaS has not tested - or " +
				"other-major - a line the template says nothing about, whose data layout or settings may " +
				"differ: warn before using one. newer says it is a later release than currentTag. It reads the " +
				"registry, so it is slower than the other store tools; truncated says there were more tags " +
				"than were read.",
			paths: map[under]string{underNothing: "/app-templates/{template}/image-tags"},
			item:  &pathItem{arg: argTemplate, description: descTemplate},
			query: &apptemplatedto.GetAppTemplateImageTagsReq{},
			params: map[string]string{
				"version": "the template's version whose image to read; the default version when not given",
				"variant": "the variant whose image to read; the default variant when not given",
			},
			answer: func() any { return &apptemplatedto.GetAppTemplateImageTagsResp{} },
		},
		{
			name: "list_sched_jobs", title: "List scheduled jobs",
			description: "GET /settings/sched-jobs, or /projects/{project}/{env}/apps/{app}/sched-jobs when an " +
				"app is given: scheduled jobs, each with its type, schedule, next runs, retries, timeout and " +
				"command. Given no app, the installation's own jobs - backups, cleanups, certificate renewal. " +
				"list_tasks with type task:sched-job-exec and targetId the job's id shows how its runs went.",
			paths: map[under]string{underNothing: "/settings/sched-jobs", underApp: "/sched-jobs"},
			query: &schedjobdto.ListSchedJobReq{},
			params: pagingParams(map[string]string{
				paramSearch: descSearchName,
				paramStatus: "only jobs in these states: " + settingStatuses,
				argKind: "only jobs of these types: " + statusValues(base.AllSchedJobTypes) + ". An app's jobs " +
					"are container-command: a command run in the app's container",
			}),
			answer: func() any { return &schedjobdto.ListSchedJobResp{} },
		},
		listEnvLinkTargetsEndpoint,
		getEnvSelfSuggestionsEndpoint,
	}, projectEndpoints()...)
}
