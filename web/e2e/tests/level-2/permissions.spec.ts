import type { APIRequestContext } from "@playwright/test";

import { type App, appPath, createApp, createProject, deleteProject, deployImage } from "../../support/api";
import { BUSYBOX, deployed } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { PASSWORD, REFUSED, devOnly, signedInAs } from "../../support/users";

test.describe.configure({ timeout: 300_000 });

// tasksOf are the IDs of the app's tasks, as an admin lists them.
async function tasksOf(api: APIRequestContext, app: App): Promise<string[]> {
    const res = await api.get(`${appPath(app)}/tasks`);
    expect(res.ok(), `the tasks of ${app.name}: ${res.status()}`).toBe(true);
    return (((await res.json()) as { data: { id: string }[] | null }).data ?? []).map(task => task.id);
}

// A member reads the tasks of the project given, and no other's: its filters -
// by project, env, app - keep to the project, and to the envs given; a task of
// another env is not read, nor its log, at an address of the project's.
test("a member's tasks of the project show no other project's, nor another env's", async ({
    page,
    api,
    cleanup,
    browser,
}) => {
    const { project, dev, prod, member } = await devOnly(page, browser, api, cleanup, "task-reader");
    const other = await createProject(api, e2eName("task-other"));
    cleanup(() => deleteProject(api, other.id));
    const otherApp = await createApp(api, other, "web");
    for (const app of [dev, prod, otherApp]) {
        await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    }
    for (const app of [dev, prod, otherApp]) {
        await deployed(api, app);
    }
    const [devTasks, prodTasks, otherTasks] = [
        await tasksOf(api, dev),
        await tasksOf(api, prod),
        await tasksOf(api, otherApp),
    ];
    expect(prodTasks.length && otherTasks.length && devTasks.length, "each app has a task").toBeTruthy();

    const memberApi = await signedInAs(member.username, PASSWORD);
    try {
        const own = `projects/${project.id}/tasks`;
        const read = async (path: string, params: Record<string, string> = {}): Promise<string[]> => {
            const res = await memberApi.get(path, { params: { pageLimit: 200, ...params } });
            if (REFUSED.includes(res.status())) return [];
            expect(res.ok(), `${path} ${JSON.stringify(params)}: ${res.status()}`).toBe(true);
            return (((await res.json()) as { data: { id: string }[] | null }).data ?? []).map(task => task.id);
        };
        const views = {
            "the project's": await read(own),
            "filtered to another project": await read(own, { projectId: other.id }),
            "filtered to another project's app": await read(own, { appId: otherApp.id }),
            "filtered to production": await read(own, { projectEnvId: `${project.id}:prod` }),
            "filtered to the production app": await read(own, { appId: prod.id }),
            "development's, filtered to another project": await read(`projects/${project.id}/${dev.env}/tasks`, {
                projectId: other.id,
            }),
        };
        for (const [what, ids] of Object.entries(views)) {
            expect
                .soft(
                    ids.filter(id => otherTasks.includes(id)),
                    `${what}: another project's tasks`,
                )
                .toEqual([]);
            expect
                .soft(
                    ids.filter(id => prodTasks.includes(id)),
                    `${what}: production's tasks`,
                )
                .toEqual([]);
        }
        expect(views["the project's"], "the development app's tasks are there").toEqual(
            expect.arrayContaining(devTasks),
        );

        const [prodTask] = prodTasks;
        const attempts = {
            "reading production's task": await memberApi.get(`${own}/${prodTask}`),
            "reading its log": await memberApi.get(`${own}/${prodTask}/logs`),
            "reading it through development": await memberApi.get(
                `projects/${project.id}/${dev.env}/tasks/${prodTask}`,
            ),
        };
        for (const [what, res] of Object.entries(attempts)) {
            expect.soft(REFUSED, `the member is refused ${what}: ${res.status()}`).toContain(res.status());
        }
    } finally {
        await memberApi.dispose();
    }
});
