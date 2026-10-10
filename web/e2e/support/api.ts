import { type APIRequestContext, type APIResponse, expect } from "@playwright/test";

// Helpers over the REST API, for what a test needs made before it starts and
// removed after it ends. Each checks the answer: a test must not go on from a
// setup that failed.

export interface Project {
    id: string;
    name: string;
    key: string;
    envs?: { name: string }[];
}

async function ok(res: APIResponse, doing: string): Promise<unknown> {
    expect(res.ok(), `${doing}: ${res.status()} ${await res.text()}`).toBe(true);
    const body = (await res.text()).trim();
    return body === "" ? undefined : (JSON.parse(body) as unknown);
}

// findProject is the project named exactly so, if there is one.
export async function findProject(api: APIRequestContext, name: string): Promise<Project | undefined> {
    const body = (await ok(
        await api.get("projects", { params: { search: name } }),
        `listing projects named ${name}`,
    )) as { data: Project[] };
    return body.data.find(p => p.name === name);
}

// createProject makes an active project with the environments the dialog
// starts with.
export async function createProject(api: APIRequestContext, name: string): Promise<Project> {
    const envs = [
        { name: "development", color: "#a855f7" },
        { name: "production", color: "#84cc16" },
    ];
    await ok(await api.post("projects", { data: { name, status: "active", envs } }), `creating project ${name}`);
    const project = await findProject(api, name);
    expect(project, `project ${name} is listed once made`).toBeDefined();
    return project as Project;
}

// createSettingAt makes a setting at a path - an app's secrets, an
// environment's config files - and answers its id.
export async function createSettingAt(api: APIRequestContext, path: string, data: object): Promise<string> {
    const body = (await ok(await api.post(path, { data }), `creating ${path}`)) as { data: { id: string } };
    return body.data.id;
}

// settingIdNamed is the id of the setting of that name listed at a path.
export async function settingIdNamed(api: APIRequestContext, path: string, name: string): Promise<string> {
    const res = await api.get(path);
    expect(res.ok(), `listing ${path}: ${res.status()}`).toBe(true);
    const found = ((await res.json()) as { data: { id: string; name: string }[] }).data.find(s => s.name === name);
    expect(found, `${name} at ${path}`).toBeDefined();
    return found?.id ?? "";
}

// deleteProject removes a project and its stored data; one already gone is
// fine.
export async function deleteProject(api: APIRequestContext, id: string): Promise<void> {
    const res = await api.delete(`projects/${id}`, { params: { removeStorage: true } });
    if (res.status() === 404) return;
    await ok(res, `deleting project ${id}`);
}

// deleteApiKeysNamed removes the current user's API keys of that name.
export async function deleteApiKeysNamed(api: APIRequestContext, name: string): Promise<void> {
    const body = (await ok(await api.get("users/current/settings/api-keys"), "listing API keys")) as {
        data: { id: string; name: string }[];
    };
    for (const key of body.data.filter(k => k.name === name)) {
        await ok(await api.delete(`users/current/settings/api-keys/${key.id}`), `deleting API key ${name}`);
    }
}

// deleteSettingsNamed removes the global settings of a kind - basic-auth,
// ssh-keys, repo-webhooks... - that carry the name.
export async function deleteSettingsNamed(api: APIRequestContext, kind: string, name: string): Promise<void> {
    const body = (await ok(await api.get(`settings/${kind}`), `listing ${kind}`)) as {
        data: { id: string; name: string }[];
    };
    for (const item of body.data.filter(i => i.name === name)) {
        await ok(await api.delete(`settings/${kind}/${item.id}`), `deleting ${kind} ${name}`);
    }
}

const CLEANUP_FIELDS = [
    "status",
    "schedule",
    "dbObjectRetention",
    "clusterCleanup",
    "cacheCleanup",
    "fileCleanup",
    "systemAppsSync",
    "notification",
] as const;

type Settings = Record<string, unknown>;

// cleanupSettings are the system cleanup's settings as a save sends them.
export async function cleanupSettings(api: APIRequestContext): Promise<Settings> {
    const body = (await ok(await api.get("system/settings/cleanup"), "reading cleanup settings")) as {
        data: Settings;
    };
    return Object.fromEntries(CLEANUP_FIELDS.map(field => [field, body.data[field]]));
}

// restoreCleanupSettings saves settings read before, over whatever a test saved
// since.
export async function restoreCleanupSettings(api: APIRequestContext, saved: Settings): Promise<void> {
    const current = (await ok(await api.get("system/settings/cleanup"), "reading cleanup settings")) as {
        data: { updateVer: number };
    };
    await ok(
        await api.put("system/settings/cleanup", { data: { ...saved, updateVer: current.data.updateVer } }),
        "restoring cleanup settings",
    );
}

// deleteUsersByEmail removes the users, invited or signed up, with that email.
export async function deleteUsersByEmail(api: APIRequestContext, email: string): Promise<void> {
    const body = (await ok(await api.get("users", { params: { search: email } }), `listing users ${email}`)) as {
        data: { id: string; email: string }[];
    };
    for (const user of body.data.filter(u => u.email === email)) {
        await ok(await api.delete(`users/${user.id}`), `deleting user ${email}`);
    }
}

// created is the id of what a POST made.
export async function created(api: APIRequestContext, path: string, data: unknown): Promise<string> {
    const body = (await ok(await api.post(path, { data }), `POST ${path}`)) as { data: { id: string } };
    return body.data.id;
}

export interface App {
    id: string;
    name: string;
    projectId: string;
    env: string;
}

/** appPath is an app's path under the API. */
export function appPath(app: App): string {
    return `projects/${app.projectId}/${app.env}/apps/${app.id}`;
}

// createApp makes an app in a project's environment. HivePaaS gives it a
// service at once, on a placeholder image, until it is deployed.
export async function createApp(
    api: APIRequestContext,
    project: Pick<Project, "id">,
    name: string,
    env = "development",
): Promise<App> {
    const body = (await ok(
        await api.post(`projects/${project.id}/${env}/apps`, {
            data: { name, env, note: "", tags: [], status: "active" },
        }),
        `creating app ${name}`,
    )) as { data: { id: string } };
    return { id: body.data.id, name, projectId: project.id, env };
}

// deployImage sets an app to run an image - with a command, when given, and a
// command run in its running container before - which deploys it; it answers
// the deployment's id.
export async function deployImage(
    api: APIRequestContext,
    app: App,
    image: string,
    command = "",
    preDeploymentCommand = "",
): Promise<string> {
    const current = (await ok(await api.get(`${appPath(app)}/deployment-settings`), "reading deployment settings")) as {
        data: { updateVer: number };
    };
    const body = (await ok(
        await api.put(`${appPath(app)}/deployment-settings`, {
            data: {
                entrypoint: "",
                command,
                workingDir: "",
                preDeploymentCommand,
                postDeploymentCommand: "",
                notification: { successUseDefault: true, failureUseDefault: true },
                activeMethod: "image",
                imageSource: { image, registryAuth: { id: "" } },
                updateVer: current.data.updateVer,
            },
        }),
        `deploying ${image} to ${app.name}`,
    )) as { data: { deploymentId: string } };
    return body.data.deploymentId;
}

export interface Deployment {
    id: string;
    status: string;
}

// latestDeployment is an app's newest deployment, if it has one.
export async function latestDeployment(api: APIRequestContext, app: App): Promise<Deployment | undefined> {
    const body = (await ok(await api.get(`${appPath(app)}/deployments`), "listing deployments")) as {
        data: Deployment[];
    };
    return body.data[0];
}

// findApp is the app of that id in its environment, if it is still there.
export async function findApp(api: APIRequestContext, app: App): Promise<App | undefined> {
    const body = (await ok(await api.get(`projects/${app.projectId}/${app.env}/apps`), "listing apps")) as {
        data: { id: string }[];
    };
    return body.data.some(a => a.id === app.id) ? app : undefined;
}

export interface EnvVar {
    key: string;
    value: string;
    isLiteral: boolean;
}

// runtimeEnvVars are the variables an app sets for itself at run time, as
// saved.
export async function runtimeEnvVars(api: APIRequestContext, app: App): Promise<EnvVar[]> {
    const body = (await ok(await api.get(`${appPath(app)}/env-vars`), "reading env variables")) as {
        data: { runtimeEnvVars: EnvVar[] | null };
    };
    return body.data.runtimeEnvVars ?? [];
}

// createVolume makes a docker volume on the node HivePaaS runs on, for apps to
// mount - inheritable, or no project sees it; it answers its id. Given a
// directory, the volume is that directory of the node, bound.
export async function createVolume(api: APIRequestContext, name: string, directory?: string): Promise<string> {
    const body = (await ok(
        await api.post("cluster/volumes", {
            data: {
                name,
                driver: "local",
                nodeId: "current",
                inheritable: true,
                ...(directory ? { bindOptions: { directory } } : {}),
            },
        }),
        `creating volume ${name}`,
    )) as { data: { id: string } };
    return body.data.id;
}

// deleteVolume removes a volume; one already gone is fine. Docker refuses one
// still in use: a project's containers go a little after the project does.
export async function deleteVolume(api: APIRequestContext, id: string): Promise<void> {
    let res = await api.delete(`cluster/volumes/${id}`);
    for (let tries = 0; res.status() === 409 && tries < 30; tries++) {
        await new Promise(resolve => setTimeout(resolve, 2_000));
        res = await api.delete(`cluster/volumes/${id}`);
    }
    if (res.status() === 404) return;
    await ok(res, `deleting volume ${id}`);
}

// setRuntimeEnvVars replaces the variables an app sets for itself at run time.
export async function setRuntimeEnvVars(api: APIRequestContext, app: App, vars: EnvVar[]): Promise<void> {
    const current = (await ok(await api.get(`${appPath(app)}/env-vars`), "reading env variables")) as {
        data: { updateVer: number };
    };
    await ok(
        await api.put(`${appPath(app)}/env-vars`, {
            data: {
                updateVer: current.data.updateVer,
                runtimeEnvVars: vars,
                buildtimeEnvVars: [],
                sharedEnvVars: [],
            },
        }),
        `setting env variables of ${app.name}`,
    );
}

// findAppNamed is the app of that name in an app's environment, if there is
// one.
export async function findAppNamed(api: APIRequestContext, near: App, name: string): Promise<App | undefined> {
    const body = (await ok(await api.get(`projects/${near.projectId}/${near.env}/apps`), "listing apps")) as {
        data: { id: string; name: string }[];
    };
    const found = body.data.find(a => a.name === name);
    return found ? { id: found.id, name, projectId: near.projectId, env: near.env } : undefined;
}

export interface SystemProject {
    id: string;
    updateVer: number;
    envs: { id: string; name: string; updateVer: number }[];
}

// systemProject is the project HivePaaS runs in, which the projects list leaves
// out, with its environments.
export async function systemProject(api: APIRequestContext): Promise<SystemProject> {
    const found = (await ok(await api.get("system/hivepaas/project"), "finding the HivePaaS project")) as {
        data: { id: string };
    };
    const body = (await ok(await api.get(`projects/${found.data.id}`), "reading the HivePaaS project")) as {
        data: SystemProject;
    };
    return body.data;
}

export interface SystemApp extends App {
    key: string;
    updateVer: number;
}

// systemApp is the app of the HivePaaS project with that key, in whichever of
// its environments it is.
export async function systemApp(api: APIRequestContext, project: SystemProject, key: string): Promise<SystemApp> {
    for (const env of project.envs) {
        const body = (await ok(
            await api.get(`projects/${project.id}/${env.name}/apps`),
            `listing the HivePaaS apps of ${env.name}`,
        )) as { data: { id: string; name: string; key: string; updateVer: number }[] };
        const app = body.data.find(a => a.key === key);
        if (app) return { ...app, projectId: project.id, env: env.name };
    }
    throw new Error(`the HivePaaS project has no app ${key}`);
}

// createBasicAuth makes credentials of the installation's, open to every
// project, for a domain to ask its visitors for.
export async function createBasicAuth(
    api: APIRequestContext,
    name: string,
    username: string,
    password: string,
): Promise<void> {
    await ok(
        await api.post("settings/basic-auth", { data: { name, username, password, inheritable: true } }),
        `creating basic auth ${name}`,
    );
}

// exposeApp opens an app to the internet at a domain - on its container's port
// 80, HTTPS forced, unless told otherwise - what Routing Settings saves for a
// new domain.
export async function exposeApp(
    api: APIRequestContext,
    app: App,
    domain: string,
    { port = 80, forceHttps = true }: { port?: number; forceHttps?: boolean } = {},
): Promise<void> {
    const current = (await ok(await api.get(`${appPath(app)}/routing-settings`), "reading routing settings")) as {
        data: { updateVer: number };
    };
    await ok(
        await api.put(`${appPath(app)}/routing-settings`, {
            data: {
                port,
                exposePublicly: true,
                domains: [{ enabled: true, domain, protocol: "http", forceHttps }],
                updateVer: current.data.updateVer,
            },
        }),
        `exposing ${app.name} at ${domain}`,
    );
}

// createJob makes a scheduled job of the app's running the command in its
// container, with no schedule - run by hand, or as a step of a sequence - and
// answers its id. `more` sets more of the job: a timeout, retries.
export async function createJob(
    api: APIRequestContext,
    app: App,
    name: string,
    command: string,
    more: Record<string, unknown> = {},
): Promise<string> {
    const body = (await ok(
        await api.post(`${appPath(app)}/sched-jobs`, {
            data: { name, jobType: "container-command", app: { id: app.id }, command: { command }, ...more },
        }),
        `creating job ${name}`,
    )) as { data: { id: string } };
    return body.data.id;
}

// runJob runs a job of the app now, waits for the run to end, and answers how
// it ended and what it logged.
export async function runJob(
    api: APIRequestContext,
    app: App,
    jobId: string,
): Promise<{ status: string; log: string }> {
    const body = (await ok(
        await api.post(`${appPath(app)}/sched-jobs/${jobId}/exec`, { data: {} }),
        `running job ${jobId}`,
    )) as { data: { task: { id: string } } };
    const task = `${appPath(app)}/tasks/${body.data.task.id}`;
    let status = "";
    await expect
        .poll(
            async () => {
                status = ((await (await api.get(task)).json()) as { data: { status: string } }).data.status;
                return ["done", "failed", "canceled"].includes(status);
            },
            { timeout: 120_000, intervals: [2_000] },
        )
        .toBe(true);
    const logs = (await (await api.get(`${task}/logs`)).json()) as { data: { logs: { data: string }[] } };
    return { status, log: logs.data.logs.map(l => l.data).join("") };
}

export interface StorageMount {
    target: string;
    // The app whose directory the mount reaches, when it is not the app's own.
    sourceApp?: { appId: string };
}

// storageMounts are the mounts an app's Persistent Storage lists.
export async function storageMounts(api: APIRequestContext, app: App): Promise<StorageMount[]> {
    const body = (await ok(
        await api.get(`${appPath(app)}/storage-settings`, { params: { getMounts: true } }),
        `reading the storage of ${app.name}`,
    )) as { data: { mounts?: StorageMount[] } };
    return body.data.mounts ?? [];
}
