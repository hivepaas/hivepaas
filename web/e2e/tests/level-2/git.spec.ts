import type { APIRequestContext, Page } from "@playwright/test";
import { createHash } from "node:crypto";

import { type App, appPath, deleteSettingsNamed, exposeApp } from "../../support/api";
import { appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import {
    GIT_TOKEN,
    REPOS,
    SHOP_COMMITS,
    buildFromRepo,
    createSetting,
    gitSshKey,
    saveRepoSource,
} from "../../support/git";
import { domainFor, visit } from "../../support/routing";

// A build clones the repository and builds its image before it deploys: each
// test has room.
test.describe.configure({ timeout: 300_000 });

interface SavedRepoSource {
    repoURL: string;
    repoRef: string;
    autoDeploy: boolean;
    dockerfile: { source: string };
    credentials: { id: string };
}

// savedRepoSource is the repository the app is built from, as saved.
async function savedRepoSource(api: APIRequestContext, app: App): Promise<SavedRepoSource> {
    const res = await api.get(`${appPath(app)}/deployment-settings`);
    expect(res.ok(), `reading deployment settings: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data: { repoSource: SavedRepoSource } }).data.repoSource;
}

// gitSource opens the app's Deployment Settings and picks Git Source, with the
// repository's address and a branch.
async function gitSource(page: Page, app: App, repoURL: string, branch: string): Promise<void> {
    await page.goto(appPage(app, "deployment-settings"));
    await page.getByRole("button", { name: /^Git Source/ }).click();
    await page
        .getByRole("group", { name: /^Git Repository/ })
        .getByRole("textbox")
        .fill(repoURL);
    await page
        .getByRole("group", { name: /^Branch/ })
        .getByRole("textbox")
        .fill(branch);
}

// The repository, the branch and the Dockerfile picked in Deployment Settings
// are what the deployment builds: the branch's commit, run. Deploy on Push
// turned off is saved off.
test("an app set in Deployment Settings to a repository's branch is built from it", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "git-ui");
    await gitSource(page, app, REPOS.shop, "develop");
    await page.getByRole("checkbox", { name: "Deploy when the branch is pushed" }).setChecked(false);
    await page.getByRole("button", { name: "Deploy", exact: true }).click();

    await deployed(api, app);
    await expectLogs(page, app, "built-from-commit-1");
    const deployment = (await (await api.get(`${appPath(app)}/deployments`)).json()) as {
        data: { output?: { commitHash?: string } }[];
    };
    expect(deployment.data[0]?.output?.commitHash, "the branch's commit is built").toBe(SHOP_COMMITS.first);
    const saved = await savedRepoSource(api, app);
    expect(saved.repoURL).toBe(REPOS.shop);
    expect(saved.autoDeploy, "Deploy on Push is saved off").toBe(false);
});

// A repository with no Dockerfile - a page of HTML - is built with one
// HivePaaS writes for it: a static web server, serving the page on port 8080.
test("a repository with no Dockerfile is built with one generated for it, and serves", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "git-auto");
    await gitSource(page, app, REPOS.site, "main");
    await page.getByRole("group", { name: "Dockerfile Source" }).getByRole("tab", { name: "Auto-Generate" }).click();
    await page.getByRole("button", { name: "Deploy", exact: true }).click();

    await deployed(api, app);
    expect((await savedRepoSource(api, app)).dockerfile.source).toBe("auto");
    const domain = domainFor("git-auto");
    await exposeApp(api, app, domain, { port: 8080, forceHttps: false });
    await visit(page, `http://${domain}/`);
    await expect(page.getByRole("heading", { name: "site-from-git" })).toBeVisible();
});

// A build-time variable is the build's argument; one that uses a secret is
// given to the build as a BuildKit secret - the Dockerfile reads it from a
// file, and the deployment's log does not show it.
test("build-time variables reach the build, a secret one as a BuildKit secret", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "git-args");
    const greeting = e2eName("greeting");
    const token = e2eName("token");
    const secret = await api.post(`${appPath(app)}/secrets`, {
        data: { key: "BUILD_TOKEN", value: token, base64: false },
    });
    expect(secret.ok(), `creating the secret: ${secret.status()} ${await secret.text()}`).toBe(true);

    await page.goto(appPage(app, "env-variables"));
    const buildtime = page.getByRole("region", { name: "Buildtime Env Variables", exact: true });
    for (const [key, value] of [
        ["GREETING", greeting],
        ["BUILD_TOKEN", "${secrets.BUILD_TOKEN}"],
    ] as const) {
        await buildtime.getByRole("button", { name: "Add" }).click();
        await buildtime.getByRole("group", { name: "New variable" }).getByRole("textbox", { name: "Key" }).fill(key);
        await buildtime
            .getByRole("group", { name: key, exact: true })
            .getByRole("textbox", { name: "Value" })
            .fill(value);
    }
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();

    const id = await buildFromRepo(api, app, { repoURL: REPOS.args, branch: "main" });
    await deployed(api, app, id);
    await expectLogs(page, app, `greeting=${greeting}`);
    // The Dockerfile prints what it read of the secret as a hash: the value
    // itself is never printed, by the build or the container.
    const hash = createHash("sha256").update(token).digest("hex").slice(0, 12);
    await expectLogs(page, app, `token=${hash}`);

    const logs = await api.get(`${appPath(app)}/deployments/${id}/logs`);
    expect(logs.ok(), `reading the deployment's log: ${logs.status()}`).toBe(true);
    const text = await logs.text();
    expect(text, "the build's step shows the argument it was given").toContain(`greeting=${greeting}`);
    expect(text, "the secret is not in the deployment's log").not.toContain(token);
});

// A private repository is refused without credentials; with an access token,
// picked in Deployment Settings, HivePaaS clones it over HTTPS.
test("a private repository is built over HTTPS with an access token picked in Deployment Settings", async ({
    page,
    api,
    cleanup,
}) => {
    // Made before the app: the app, which uses it, goes first.
    const name = e2eName("git-token");
    cleanup(() => deleteSettingsNamed(api, "access-tokens", name));
    await createSetting(api, "access-tokens", { kind: "gitea", name, token: GIT_TOKEN, user: "", baseURL: "" });
    const app = await appIn(api, cleanup, "git-https");

    // Without credentials: not found, as a private repository is to a stranger.
    const refused = await saveRepoSource(api, app, { repoURL: REPOS.privateHttps, branch: "main" });
    expect(refused.ok(), "a private repository without credentials is refused").toBe(false);
    expect(((await refused.json()) as { code: string }).code).toBe("ERR_REPO_NOT_FOUND");

    await gitSource(page, app, REPOS.privateHttps, "main");
    await page.getByRole("group", { name: "Git Credentials" }).getByRole("combobox").click();
    await page.getByRole("option", { name: new RegExp(name) }).click();
    await page.getByRole("button", { name: "Deploy", exact: true }).click();

    await deployed(api, app);
    await expectLogs(page, app, "built-from-private");
});

// The same private repository over SSH, with a key of the installation's.
test("a private repository is built over SSH with a key", async ({ page, api, cleanup }) => {
    const name = e2eName("git-key");
    cleanup(() => deleteSettingsNamed(api, "ssh-keys", name));
    const key = await createSetting(api, "ssh-keys", {
        kind: "git",
        name,
        keyType: "ed25519",
        publicKey: "",
        privateKey: gitSshKey(),
        passphrase: "",
    });
    const app = await appIn(api, cleanup, "git-ssh");

    const id = await buildFromRepo(api, app, { repoURL: REPOS.privateSsh, branch: "main", credentials: key });
    await deployed(api, app, id);
    await expectLogs(page, app, "built-from-private");
});
