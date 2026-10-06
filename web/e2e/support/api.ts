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
