import type { APIRequestContext } from "@playwright/test";
import { createHmac, randomBytes } from "node:crypto";

import { type App, appPath, deleteSettingsNamed } from "../../support/api";
import { DEPLOYED, appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// The repository env/up.sh serves inside dind, where the backend reaches it. A
// Git host names it by its page's address. Its main has two commits, each a
// Dockerfile that prints which it is; develop is at the first. up.sh makes them
// at fixed dates: their hashes are these on every env.
const REPO_URL = "git://127.0.0.1/e2e/shop.git";
const REPO_PAGE = "http://127.0.0.1/e2e/shop";
const FIRST_COMMIT = "2010c034663e785fac6e0d81eb55232d5abfcc0b";
const SECOND_COMMIT = "860f1025bb11e0171549d86ac75c002cad5caa56";

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

// buildFromRepo sets the app to be built from a repository's branch, at a commit
// of it, with the Dockerfile found there, which deploys it.
async function buildFromRepo(
    api: APIRequestContext,
    app: App,
    repoURL: string,
    branch: string,
    commitHash: string,
): Promise<void> {
    const current = await api.get(`${appPath(app)}/deployment-settings`);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    const res = await api.put(`${appPath(app)}/deployment-settings`, {
        data: {
            entrypoint: "",
            command: "",
            workingDir: "",
            preDeploymentCommand: "",
            postDeploymentCommand: "",
            notification: { successUseDefault: true, failureUseDefault: true },
            activeMethod: "repo",
            repoSource: {
                repoType: "git",
                repoURL,
                repoRef: branch,
                commitHash,
                credentials: { id: "" },
                dockerfile: { source: "manual", path: "Dockerfile" },
                pushToRegistry: { id: "" },
            },
            updateVer,
        },
    });
    expect(res.ok(), `building ${app.name} from ${repoURL}: ${res.status()} ${await res.text()}`).toBe(true);
}

// push is what GitHub sends of a push of a commit to a branch, signed with the
// webhook's secret, as it signs it.
function push(branch: string, commit: string, secret: string) {
    const body = JSON.stringify({
        ref: `refs/heads/${branch}`,
        after: commit,
        repository: { html_url: REPO_PAGE, clone_url: REPO_URL, name: "shop", full_name: "e2e/shop" },
    });
    const signature = createHmac("sha256", secret).update(body).digest("hex");
    return {
        headers: {
            "Content-Type": "application/json",
            "X-GitHub-Event": "push",
            "X-GitHub-Delivery": randomBytes(16).toString("hex"),
            "X-Hub-Signature-256": `sha256=${signature}`,
        },
        data: body,
    };
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
    const name = e2eName("webhook");
    const secret = randomBytes(16).toString("hex");
    cleanup(() => deleteSettingsNamed(api, "repo-webhooks", name));
    const created = await api.post("settings/repo-webhooks", { data: { name, kind: "github", secret } });
    expect(created.ok(), `creating the webhook: ${created.status()} ${await created.text()}`).toBe(true);
    const webhook = ((await created.json()) as { data: { id: string } }).data;

    // Built from main at its first commit, and run.
    const app = await appIn(api, cleanup, "webhook");
    await buildFromRepo(api, app, REPO_URL, "main", FIRST_COMMIT);
    await deployed(api, app);
    await expectLogs(page, app, "built-from-commit-1");

    // Another branch: nothing.
    let res = await api.post(`webhooks/${webhook.id}`, push("develop", FIRST_COMMIT, secret));
    expect(res.ok(), `a push to develop: ${res.status()} ${await res.text()}`).toBe(true);
    // Not signed with the secret: refused, and nothing.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SECOND_COMMIT, "not-the-secret"));
    expect(res.status(), "a push signed with another secret is refused").toBe(401);
    expect(((await res.json()) as { code: string }).code).toBe("ERR_WEBHOOK_UNVERIFIED");
    expect(await deploymentsOf(api, app), "neither deploys").toHaveLength(1);

    // The second commit pushed to main: built and deployed, by the webhook.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SECOND_COMMIT, secret));
    expect(res.ok(), `a push to main: ${res.status()} ${await res.text()}`).toBe(true);
    await expect.poll(async () => (await deploymentsOf(api, app)).length, { timeout: 30_000 }).toBe(2);
    expect((await deploymentsOf(api, app))[0]?.trigger?.source).toBe("repo-webhook");
    await expect.poll(async () => (await deploymentsOf(api, app))[0]?.status, DEPLOYED).toBe("done");
    expect((await deploymentsOf(api, app))[0]?.output?.commitHash).toBe(SECOND_COMMIT);
    await expectLogs(page, app, "built-from-commit-2");

    // The same push delivered again: nothing more.
    res = await api.post(`webhooks/${webhook.id}`, push("main", SECOND_COMMIT, secret));
    expect(res.ok(), `the push again: ${res.status()} ${await res.text()}`).toBe(true);
    expect(await deploymentsOf(api, app), "a push is deployed once").toHaveLength(2);

    await page.goto(appPage(app, "deployments"));
    await expect(page.getByRole("button", { name: /Webhook/ })).toHaveCount(1);
});
