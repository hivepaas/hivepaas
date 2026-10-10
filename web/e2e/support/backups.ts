import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, createVolume, created, deleteSettingsNamed, deleteVolume } from "./api";
import { type Cleanup, e2eName, expect } from "./fixtures";

export const REPO_PASSWORD = "E2e-backup-Pass1";

export interface Repo {
    id: string;
    name: string;
}

// storeVolume makes the volume a backup repository is kept on, a directory of
// the node, named as the repository is to be; it goes once the test ends,
// after the repository on it.
export async function storeVolume(
    api: APIRequestContext,
    cleanup: Cleanup,
    label: string,
): Promise<{ id: string; name: string }> {
    const name = e2eName(label);
    const id = await createVolume(api, name, `/hp/${name}`);
    cleanup(() => deleteVolume(api, id));
    cleanup(() => deleteSettingsNamed(api, "backup-repos", name));
    return { id, name };
}

// volumeRepo makes a backup repository on a volume of its own, available in
// every project; both go once the test ends.
export async function volumeRepo(api: APIRequestContext, cleanup: Cleanup, label: string): Promise<Repo> {
    const store = await storeVolume(api, cleanup, label);
    const id = await created(api, "settings/backup-repos", {
        name: store.name,
        engine: "kopia",
        volume: { id: store.id },
        password: REPO_PASSWORD,
        inheritable: true,
    });
    return { id, name: store.name };
}

// volumeBackupJob makes a data backup job of the app: what the app has of the
// volume, into the repository.
export async function volumeBackupJob(
    api: APIRequestContext,
    app: App,
    name: string,
    volumeId: string,
    repo: Repo,
): Promise<string> {
    return created(api, `${appPath(app)}/sched-jobs`, {
        name,
        jobType: "data-backup",
        app: { id: app.id },
        dataBackup: { source: "volume", sourceVolume: { id: volumeId }, targetRepository: { id: repo.id } },
    });
}

export interface Snapshot {
    id: string;
    shortId: string;
}

// snapshotsIn are the snapshots HivePaaS knows the repository holds.
export async function snapshotsIn(api: APIRequestContext, repo: Repo): Promise<Snapshot[]> {
    const res = await api.get("settings/backup-snapshots", { params: { repo: repo.id } });
    expect(res.ok(), `listing the snapshots of ${repo.name}: ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: Snapshot[] }).data;
}

// setRepoStatus makes the repository active, or not.
export async function setRepoStatus(api: APIRequestContext, repo: Repo, status: "active" | "disabled"): Promise<void> {
    const current = await api.get(`settings/backup-repos/${repo.id}`);
    expect(current.ok(), `reading ${repo.name}: ${current.status()}`).toBe(true);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    const res = await api.put(`settings/backup-repos/${repo.id}/status`, { data: { status, updateVer } });
    expect(res.ok(), `making ${repo.name} ${status}: ${res.status()} ${await res.text()}`).toBe(true);
}

// repoAction runs one of the repository's actions from its row of Backup
// Repos, and waits for it to say it is done.
export async function repoAction(page: Page, repo: Repo, action: "Repo Cleanup" | "Repo Sync"): Promise<void> {
    await page.goto("/integrations/backup-repos/");
    await page
        .getByRole("row", { name: new RegExp(repo.name) })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: action }).click();
    const done = action === "Repo Cleanup" ? "cleanup completed" : "synced";
    await expect(page.getByText(new RegExp(`Backup repository ${done} successfully`))).toBeVisible({
        timeout: 120_000,
    });
}

// startRestore opens the restore of the snapshot the job made, from its row of
// the Backup Snapshots page given.
export async function startRestore(page: Page, snapshots: string, job: string): Promise<void> {
    await page.goto(snapshots);
    await page
        .getByRole("row", { name: new RegExp(job) })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("menuitem", { name: "Restore" }).click();
    await expect(page.getByRole("heading", { name: /^Restore Snapshot/ })).toBeVisible();
}

// confirmRestore types the app's name and restores, and waits for the task to
// be done.
export async function confirmRestore(page: Page, appName: string): Promise<void> {
    await page.getByRole("group", { name: "App Name" }).getByRole("textbox").fill(appName);
    await page.getByRole("button", { name: "Restore", exact: true }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });
}
