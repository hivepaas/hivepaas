import type { APIRequestContext } from "@playwright/test";
import { createHmac, randomBytes } from "node:crypto";

import { type App, appPath, createVolume, deleteSettingsNamed, deleteVolume, storageMounts } from "../../support/api";
import { DEPLOYED, appIn, appPage, deployed, expectLogs, mountVolume } from "../../support/apps";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";
import { REPOS, SHOP_COMMITS, SHOP_PAGE, buildFromRepo } from "../../support/git";

test.describe.configure({ timeout: 300_000 });

interface Deployment {
    id: string;
    status: string;
    trigger?: { source: string };
    output?: { commitHash?: string };
}

async function deploymentsOf(api: APIRequestContext, app: App): Promise<Deployment[]> {
    const res = await api.get(`${appPath(app)}/deployments`);
    expect(res.ok(), `listing deployments: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data: Deployment[] }).data;
}

const SHOP_REPOSITORY = { html_url: SHOP_PAGE, clone_url: REPOS.shop, name: "shop", full_name: "e2e/shop" };

// delivery is a GitHub webhook's call: an event, signed with the webhook's
// secret, as GitHub signs it.
function delivery(event: string, payload: object, secret: string) {
    const body = JSON.stringify(payload);
    const signature = createHmac("sha256", secret).update(body).digest("hex");
    return {
        headers: {
            "Content-Type": "application/json",
            "X-GitHub-Event": event,
            "X-GitHub-Delivery": randomBytes(16).toString("hex"),
            "X-Hub-Signature-256": `sha256=${signature}`,
        },
        data: body,
    };
}

// push is what GitHub sends of a push of a commit to a branch.
function push(branch: string, commit: string, secret: string) {
    return delivery("push", { ref: `refs/heads/${branch}`, after: commit, repository: SHOP_REPOSITORY }, secret);
}

// prComment is what GitHub sends of a comment on pull request 7, by someone
// related to the repository as association says: OWNER, NONE for a stranger.
function prComment(body: string, association: string, secret: string) {
    return delivery(
        "issue_comment",
        {
            action: "created",
            issue: { number: 7, pull_request: { url: "http://127.0.0.1/e2e/shop/pull/7" } },
            comment: { body, user: { login: "e2e-commenter" }, author_association: association },
            repository: SHOP_REPOSITORY,
        },
        secret,
    );
}

// createWebhook makes a GitHub webhook of the installation's, and answers its
// id; it goes once the test ends.
async function createWebhook(api: APIRequestContext, cleanup: Cleanup, label: string, secret: string): Promise<string> {
    const name = e2eName(label);
    cleanup(() => deleteSettingsNamed(api, "repo-webhooks", name));
    const created = await api.post("settings/repo-webhooks", { data: { name, kind: "github", secret } });
    expect(created.ok(), `creating the webhook: ${created.status()} ${await created.text()}`).toBe(true);
    return ((await created.json()) as { data: { id: string } }).data.id;
}

// previewTasks counts the tasks made to create a preview of the app: a
// comment's command makes one before it answers, the preview is made later.
async function previewTasks(api: APIRequestContext, app: App): Promise<number> {
    const res = await api.get(`${appPath(app)}/tasks`);
    expect(res.ok(), `listing tasks: ${res.status()} ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: { type: string }[] }).data.filter(t => t.type === "task:app-preview").length;
}

// previewsOf are the app's previews, the apps made of its pull requests.
async function previewsOf(api: APIRequestContext, app: App): Promise<{ id: string; name: string }[]> {
    const res = await api.get(`${appPath(app)}/previews`);
    expect(res.ok(), `listing previews: ${res.status()} ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: { id: string; name: string }[] }).data;
}

// An app built from a repository's branch, and the repository's webhook, as
// GitHub calls it on a push: the commit pushed is built and deployed, marked as
// by the webhook. A push to another branch deploys nothing, nor does one not
// signed with the webhook's secret, nor a push delivered again - as GitHub does
// when it is not sure it was taken.
test("an app built from a repository deploys the commit pushed to its branch, through the webhook", async ({
    page,
    api,
    cleanup,
}) => {
    const secret = randomBytes(16).toString("hex");
    const webhook = { id: await createWebhook(api, cleanup, "webhook-push", secret) };

    // Built from main at its first commit, and run.
    const app = await appIn(api, cleanup, "webhook");
    await buildFromRepo(api, app, { repoURL: REPOS.shop, branch: "main", commitHash: SHOP_COMMITS.first });
    await deployed(api, app);
    await expectLogs(page, app, "built-from-commit-1");

    // Another branch: nothing.
    let res = await api.post(`webhooks/${webhook.id}`, push("develop", SHOP_COMMITS.first, secret));
    expect(res.ok(), `a push to develop: ${res.status()} ${await res.text()}`).toBe(true);
    // Not signed with the secret: refused, and nothing.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SHOP_COMMITS.second, "not-the-secret"));
    expect(res.status(), "a push signed with another secret is refused").toBe(401);
    expect(((await res.json()) as { code: string }).code).toBe("ERR_WEBHOOK_UNVERIFIED");
    expect(await deploymentsOf(api, app), "neither deploys").toHaveLength(1);

    // The second commit pushed to main: built and deployed, by the webhook.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SHOP_COMMITS.second, secret));
    expect(res.ok(), `a push to main: ${res.status()} ${await res.text()}`).toBe(true);
    await expect.poll(async () => (await deploymentsOf(api, app)).length, { timeout: 30_000 }).toBe(2);
    expect((await deploymentsOf(api, app))[0]?.trigger?.source).toBe("repo-webhook");
    await expect.poll(async () => (await deploymentsOf(api, app))[0]?.status, DEPLOYED).toBe("done");
    expect((await deploymentsOf(api, app))[0]?.output?.commitHash).toBe(SHOP_COMMITS.second);
    await expectLogs(page, app, "built-from-commit-2");

    // The same push delivered again: nothing more.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SHOP_COMMITS.second, secret));
    expect(res.ok(), `the push again: ${res.status()} ${await res.text()}`).toBe(true);
    expect(await deploymentsOf(api, app), "a push is deployed once").toHaveLength(2);

    await page.goto(appPage(app, "deployments"));
    await expect(page.getByRole("button", { name: /Webhook/ })).toHaveCount(1);
});

// Previews turned on in Feature Settings, with PR comments allowed: a comment
// "/hivepaas deploy" on a pull request, by someone who may write to the
// repository, makes the app's preview of it - an app of its own, pr-7, built
// from the pull request's head - and "/hivepaas cancel" removes it. A stranger's
// comment makes nothing.
test("a pull request's comment deploys a preview of the app, and another cancels it", async ({
    page,
    api,
    cleanup,
}) => {
    const secret = randomBytes(16).toString("hex");
    const webhook = await createWebhook(api, cleanup, "webhook-pr", secret);
    // The app has a volume: a preview is to have its own directory on it, not the
    // app's. Made before the project, to be removed after it.
    const volume = e2eName("preview-vol");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    // Not deployed on push: a push to main, in the test above, is not this one's.
    const app = await appIn(api, cleanup, "preview");
    await mountVolume(page, app, volume, "/data");
    await buildFromRepo(api, app, {
        repoURL: REPOS.shop,
        branch: "main",
        commitHash: SHOP_COMMITS.first,
        autoDeploy: false,
    });
    await deployed(api, app);

    await page.goto(appPage(app, "feature-settings"));
    const previews = page.getByText("App Preview", { exact: true }).locator("xpath=..");
    await previews.getByRole("group", { name: "Enabled", exact: true }).getByRole("checkbox").check();
    await previews
        .getByRole("group", { name: /^Allow PR Comments/ })
        .getByRole("checkbox")
        .check();
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await expect(page.getByText("Feature settings updated")).toBeVisible();

    // A stranger's comment: nothing.
    let res = await api.post(`webhooks/${webhook}`, prComment("/hivepaas deploy nowait", "NONE", secret));
    expect(res.ok(), `a stranger's comment: ${res.status()} ${await res.text()}`).toBe(true);
    expect(await previewTasks(api, app), "a stranger's comment makes no preview").toBe(0);

    // The owner's: the preview, built from the pull request's head. nowait skips
    // the delay previews are made after, for the pull request's other commits.
    res = await api.post(`webhooks/${webhook}`, prComment("/hivepaas deploy nowait", "OWNER", secret));
    expect(res.ok(), `the owner's comment: ${res.status()} ${await res.text()}`).toBe(true);
    expect(await previewTasks(api, app), "the owner's comment makes one").toBe(1);
    await expect.poll(async () => (await previewsOf(api, app)).map(p => p.name), DEPLOYED).toEqual(["pr-7"]);
    const [made] = await previewsOf(api, app);
    const preview = { ...app, id: made!.id, name: made!.name };
    await deployed(api, preview);
    await expectLogs(page, preview, "built-from-pull-request");
    // Its /data is a directory of its own: not the app's, which the preview's
    // code would otherwise read and write.
    const mounts = await storageMounts(api, preview);
    expect(mounts.map(m => m.target)).toEqual(["/data"]);
    expect(mounts[0]?.sourceApp, "the preview does not reach the app's data").toBeUndefined();
    await page.goto(appPage(app, "preview-deployments"));
    await expect(page.getByRole("row", { name: /pr-7/ })).toBeVisible();

    // Cancelled: gone.
    res = await api.post(`webhooks/${webhook}`, prComment("/hivepaas cancel", "OWNER", secret));
    expect(res.ok(), `the cancel: ${res.status()} ${await res.text()}`).toBe(true);
    await expect.poll(async () => previewsOf(api, app), { timeout: 60_000 }).toEqual([]);
});
