import type { APIRequestContext } from "@playwright/test";

import {
    appPath,
    createApp,
    createProject,
    createSettingAt,
    deleteProject,
    runtimeEnvVars,
    setRuntimeEnvVars,
} from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";
import { PASSWORD, REFUSED, devOnly, signedInAs } from "../../support/users";

// A member given one environment of a project reaches its apps, and none of
// the other's: not at their own address, not listed with the project's, and
// not at an address of the environment given.
test("a member given development is refused the production apps, by any address", async ({
    page,
    api,
    cleanup,
    browser,
}) => {
    const { project, dev, prod, member } = await devOnly(page, browser, api, cleanup, "dev-only");
    const greeting = { key: "GREETING", value: "production's own", isLiteral: true };
    await setRuntimeEnvVars(api, prod, [greeting]);
    const vars = (await (await api.get(`${appPath(prod)}/env-vars`)).json()) as { data: { updateVer: number } };
    const envSecret = await createSettingAt(api, `projects/${project.id}/production/secrets`, {
        key: "E2E_PRODUCTION",
        value: "production's secret",
    });
    const appSecret = await createSettingAt(api, `${appPath(prod)}/secrets`, {
        key: "E2E_API",
        value: "the api's secret",
    });

    const memberApi = await signedInAs(member.username, PASSWORD);
    try {
        expect((await memberApi.get(appPath(dev))).status(), "its development app").toBe(200);
        const listed = await memberApi.get(`projects/${project.id}/apps`);
        expect(listed.status()).toBe(200);
        const ids = ((await listed.json()) as { data: { id: string }[] }).data.map(app => app.id);
        expect.soft(ids, "the project's apps it is listed").toEqual([dev.id]);

        const throughDev = `projects/${project.id}/${dev.env}/apps/${prod.id}`;
        const attempts = {
            "reading it": await memberApi.get(appPath(prod)),
            "listing production's apps": await memberApi.get(`projects/${project.id}/production/apps`),
            "reading it through development": await memberApi.get(throughDev),
            "reading its variables through development": await memberApi.get(`${throughDev}/env-vars`),
            "reading production's secret": await memberApi.get(
                `projects/${project.id}/production/secrets/${envSecret}`,
            ),
            "reading production's secret through development": await memberApi.get(
                `projects/${project.id}/${dev.env}/secrets/${envSecret}`,
            ),
            "reading its secret through development": await memberApi.get(`${throughDev}/secrets/${appSecret}`),
            "reading its secret through its development neighbour": await memberApi.get(
                `${appPath(dev)}/secrets/${appSecret}`,
            ),
            // Refused or not, the body is not one of clone settings: nothing is saved.
            "setting how it clones through development": await memberApi.put(`${throughDev}/clone-settings`, {
                data: {},
            }),
            "saving its variables through development": await memberApi.put(`${throughDev}/env-vars`, {
                data: {
                    updateVer: vars.data.updateVer,
                    runtimeEnvVars: [],
                    buildtimeEnvVars: [],
                    sharedEnvVars: [],
                },
            }),
        };
        for (const [what, res] of Object.entries(attempts)) {
            expect.soft(REFUSED, `the member is refused ${what}: ${res.status()}`).toContain(res.status());
        }
    } finally {
        await memberApi.dispose();
    }
    expect(await runtimeEnvVars(api, prod), "production's variables are as they were").toEqual([
        expect.objectContaining(greeting),
    ]);
});

interface UserDetails {
    id: string;
    username: string;
    email: string;
    fullName: string;
    role: string;
    status: string;
    securityOption: string;
    accessExpireAt: string | null;
    moduleAccesses: { id: string }[] | null;
    capabilities: string[] | null;
    projectAccesses:
        { project: { id: string }; envAccesses: { id: string; access: Record<string, boolean> }[] }[] | null;
}

async function userDetails(api: APIRequestContext, id: string): Promise<UserDetails> {
    const res = await api.get(`users/${id}`, { params: { getAccesses: true } });
    expect(res.ok(), `reading the user ${id}: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data: UserDetails }).data;
}

// A member changes nothing of what an admin decides about them - their role,
// what they may reach, until when - whatever they send for their own account.
test("a member gives themselves nothing an admin has not", async ({ page, api, cleanup, browser }) => {
    const { project, member } = await devOnly(page, browser, api, cleanup, "self-made");
    const found = (await (await api.get("users", { params: { search: member.email } })).json()) as {
        data: { id: string; email: string }[];
    };
    const id = found.data.find(user => user.email === member.email)?.id ?? "";
    const expiry = new Date(Date.now() + 24 * 3_600_000).toISOString();
    const set = await api.put(`users/${id}`, {
        data: { username: member.username, email: member.email, accessExpireAt: expiry },
    });
    expect(set.ok(), `an admin sets the member's expiry: ${await set.text()}`).toBe(true);
    const before = await userDetails(api, id);

    const all = { read: true, execute: true, write: true, delete: true };
    const own = { username: member.username, email: member.email };
    const attempts = {
        "the admin role": { role: "admin" },
        "the Users module": { moduleAccesses: [{ id: "mod::user", access: all }] },
        "the Projects module": { moduleAccesses: [{ id: "mod::project", access: all }] },
        "revealing secrets": { capabilities: ["cap::secret::reveal"] },
        "production": {
            projectAccesses: [
                {
                    project: { id: project.id },
                    envAccesses: [
                        { id: `${project.id}:dev`, access: { ...all, delete: false } },
                        { id: `${project.id}:prod`, access: all },
                    ],
                },
            ],
        },
        "deleting in development": {
            projectAccesses: [{ project: { id: project.id }, envAccesses: [{ id: `${project.id}:dev`, access: all }] }],
        },
        "a later expiry": { accessExpireAt: new Date(Date.now() + 10 * 365 * 24 * 3_600_000).toISOString() },
    };
    const tries = [
        ...Object.entries(attempts).map(([what, more]) => ({ what, path: `users/${id}`, more })),
        { what: "a later expiry, as the current user", path: "users/current", more: attempts["a later expiry"] },
    ];
    for (const { what, path, more } of tries) {
        // Signed in for each: a user changed is signed out.
        const memberApi = await signedInAs(member.username, PASSWORD);
        try {
            const res = await memberApi.put(path, { data: { ...own, ...more } });
            expect
                .soft(REFUSED, `the member is refused ${what}: ${res.status()} ${await res.text()}`)
                .toContain(res.status());
        } finally {
            await memberApi.dispose();
        }
    }
    const after = await userDetails(api, id);
    expect(after.role).toBe("member");
    expect(after.accessExpireAt).toBe(before.accessExpireAt);
    expect(after.moduleAccesses ?? []).toEqual(before.moduleAccesses ?? []);
    expect(after.capabilities ?? []).toEqual(before.capabilities ?? []);
    expect(after.projectAccesses).toEqual(before.projectAccesses);

    // Nor do they remove their account: that is an admin's to do.
    const memberApi = await signedInAs(member.username, PASSWORD);
    try {
        const res = await memberApi.delete(`users/${id}`);
        expect.soft(REFUSED, `the member is refused deleting themselves: ${res.status()}`).toContain(res.status());
    } finally {
        await memberApi.dispose();
    }
    expect((await api.get(`users/${id}`)).status(), "the member's account is still there").toBe(200);
});

interface Entry {
    type: string;
    scopeProject?: { id: string };
    scopeApp?: { id: string };
}

// A member reads the audit log of the project given, and no other's: its
// filters - by project, env, app - keep to the project, and to the envs given.
test("a member's audit log of the project shows no other project's entries, nor another env's", async ({
    page,
    api,
    cleanup,
    browser,
}) => {
    const { project, dev, prod, member } = await devOnly(page, browser, api, cleanup, "log-reader");
    const other = await createProject(api, e2eName("log-other"));
    cleanup(() => deleteProject(api, other.id));
    const otherApp = await createApp(api, other, "web");
    const greeting = [{ key: "GREETING", value: "hello", isLiteral: true }];
    for (const app of [dev, prod, otherApp]) {
        await setRuntimeEnvVars(api, app, greeting);
    }

    const memberApi = await signedInAs(member.username, PASSWORD);
    try {
        const own = `projects/${project.id}/audit-logs`;
        const read = async (params: Record<string, string> = {}): Promise<Entry[]> => {
            const res = await memberApi.get(own, { params: { pageLimit: 200, ...params } });
            if (REFUSED.includes(res.status())) return [];
            expect(res.ok(), `${own} ${JSON.stringify(params)}: ${res.status()}`).toBe(true);
            return ((await res.json()) as { data: Entry[] | null }).data ?? [];
        };
        const views = {
            "the project's": await read(),
            "filtered to another project": await read({ projectId: other.id }),
            "filtered to another project's env": await read({ projectEnvId: `${other.id}:dev` }),
            "filtered to another project's app": await read({ appId: otherApp.id }),
            "filtered to production": await read({ projectEnvId: `${project.id}:prod` }),
            "filtered to the production app": await read({ appId: prod.id }),
        };
        for (const [what, entries] of Object.entries(views)) {
            const others = entries.filter(e => e.scopeProject && e.scopeProject.id !== project.id);
            expect.soft(others, `${what}: another project's entries`).toEqual([]);
            const production = entries.filter(e => e.scopeApp?.id === prod.id);
            expect.soft(production, `${what}: the production app's entries`).toEqual([]);
        }
        const developed = views["the project's"].filter(e => e.scopeApp?.id === dev.id);
        expect(developed.length, "the development app's entries are there").toBeGreaterThan(0);
    } finally {
        await memberApi.dispose();
    }
});
