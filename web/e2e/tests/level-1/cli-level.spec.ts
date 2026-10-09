import type { APIRequestContext } from "@playwright/test";

import { deleteProject, findProject } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

// The HivePaaS CLI says which API level it was built for in this header; the
// server refuses its writes below its own level (backend middleware/clilevel),
// so that an older CLI never writes back an object without a field it does not
// know. Every other client - the dashboard, these tests - sends no header.
const CLI_HEADER = "HivePaaS-CLI";

async function serverLevel(api: APIRequestContext): Promise<number> {
    const res = await api.get("sessions/me");
    expect(res.ok(), `reading the session: ${res.status()}`).toBe(true);
    const level = ((await res.json()) as { data: { server?: { apiLevel?: number } } }).data.server?.apiLevel;
    expect(level, "the server says its API level").toBeGreaterThanOrEqual(1);
    return level ?? 0;
}

function newProject(name: string): Record<string, unknown> {
    return { name, status: "active", envs: [{ name: "development", color: "#a855f7" }] };
}

test("a CLI built for an older API reads, and is refused its writes", async ({ api, cleanup }) => {
    const level = await serverLevel(api);
    const older = { [CLI_HEADER]: `0.0.1-e2e; api-level=${level - 1}` };

    const read = await api.get("projects", { headers: older });
    expect(read.status(), "an older CLI reads").toBe(200);
    expect(read.headers()["hivepaas-api-level"], "and is told the server's level").toBe(String(level));

    const name = e2eName("by-old-cli");
    cleanup(async () => {
        const stray = await findProject(api, name);
        if (stray) await deleteProject(api, stray.id);
    });
    const write = await api.post("projects", { data: newProject(name), headers: older });
    expect(write.status(), "an older CLI is refused a write").toBe(426);
    const refusal = (await write.json()) as { code: string; detail: string };
    expect(refusal.code).toBe("ERR_CLI_OUTDATED");
    expect(refusal.detail).toContain(`the server is at API level ${level}, the CLI at ${level - 1}`);
    expect(await findProject(api, name), "nothing was made").toBeUndefined();

    // A CLI that says no level cannot be trusted to write either.
    const unknown = await api.post("projects", { data: newProject(name), headers: { [CLI_HEADER]: "0.0.1-e2e" } });
    expect(unknown.status(), "a CLI that says no level is refused a write").toBe(426);
});

test("a CLI built for the server's API level writes", async ({ api, cleanup }) => {
    const level = await serverLevel(api);
    const name = e2eName("by-current-cli");
    cleanup(async () => {
        const made = await findProject(api, name);
        if (made) await deleteProject(api, made.id);
    });

    const write = await api.post("projects", {
        data: newProject(name),
        headers: { [CLI_HEADER]: `0.0.1-e2e; api-level=${level}` },
    });

    expect(write.ok(), `a current CLI writes: ${write.status()} ${await write.text()}`).toBe(true);
    expect(await findProject(api, name), "the project was made").toBeDefined();
});
