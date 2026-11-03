import type { APIResponse } from "@playwright/test";

import { appPath, systemApp, systemProject } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// The project HivePaaS runs in is open to an admin, from System › HivePaaS ›
// Actions, for its apps' logs. What in it runs this installation is not
// deleted, disabled or stopped by hand. A refusal missing here would take the
// installation down - one that is there to go, as level 1 runs on.

async function refused(res: APIResponse, code: string, detail: string): Promise<void> {
    const body = (await res.json()) as { code?: string; detail?: string };
    expect(res.status(), JSON.stringify(body)).toBe(403);
    expect(body.code).toBe(code);
    expect(body.detail).toContain(detail);
}

test.describe("the HivePaaS project", () => {
    test("is not deleted or disabled, nor are its environments", async ({ api }) => {
        const project = await systemProject(api);

        await refused(await api.delete(`projects/${project.id}`), "ERR_SYSTEM_PROJECT_PROTECTED", "cannot be deleted");
        await refused(
            await api.put(`projects/${project.id}/status`, {
                data: { updateVer: project.updateVer, status: "disabled" },
            }),
            "ERR_SYSTEM_PROJECT_PROTECTED",
            "cannot be disabled",
        );
        for (const env of project.envs) {
            await refused(
                await api.delete(`projects/${project.id}/${env.id}`),
                "ERR_SYSTEM_PROJECT_ENV_PROTECTED",
                "cannot be deleted",
            );
            await refused(
                await api.put(`projects/${project.id}/${env.id}/status`, {
                    data: { updateVer: env.updateVer, status: "disabled" },
                }),
                "ERR_SYSTEM_PROJECT_ENV_PROTECTED",
                "cannot be disabled",
            );
        }
    });

    test("its own apps are not deleted, disabled or stopped", async ({ api }) => {
        const traefik = await systemApp(api, await systemProject(api), "traefik");

        await refused(await api.delete(appPath(traefik)), "ERR_SYSTEM_APP_PROTECTED", "cannot be deleted");
        await refused(
            await api.put(`${appPath(traefik)}/status`, {
                data: { updateVer: traefik.updateVer, status: "disabled" },
            }),
            "ERR_SYSTEM_APP_PROTECTED",
            "cannot be disabled",
        );
        await refused(
            await api.post(`${appPath(traefik)}/running-status`, { data: { running: false } }),
            "ERR_SYSTEM_APP_PROTECTED",
            "cannot be stopped",
        );
    });
});
