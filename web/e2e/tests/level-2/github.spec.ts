import { createProject, deleteProject, exposeApp } from "../../support/api";
import { appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { saveRepoSource } from "../../support/git";
import { domainFor, visit } from "../../support/routing";

// GitHub itself: the organization hivepaas-test, which the installation's seed
// reaches through a GitHub App and an access token. Its go-app is public, its
// go-app-private private; both build a server that says Hello on 8080. Nothing
// is written to GitHub: the tests list and clone.
const GITHUB_APP = "hivepaas-test donotdelete";
const GITHUB_APP_ID = "01JAB9XED0GTXBSQDFVYAJ8WJ1";
const GITHUB_TOKEN_ID = "01JAB9XED0GTXBSQDFVYAJ8WK1";
const REPO = "https://github.com/hivepaas-test/go-app.git";
const PRIVATE_REPO = "https://github.com/hivepaas-test/go-app-private.git";

// A build fetches a Go toolchain and the modules: room for it.
test.describe.configure({ timeout: 600_000 });
test.use({ ignoreHTTPSErrors: true });

// The repository and its branch are picked from what the GitHub App lists, and
// the app built from them serves at its domain.
test("an app built from a GitHub repository picked through the GitHub App serves", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "github-public");
    await page.goto(appPage(app, "deployment-settings"));
    await page.getByRole("button", { name: /^Git Source/ }).click();
    await page.getByRole("group", { name: "Git Credentials" }).getByRole("combobox").click();
    await page.getByRole("option", { name: new RegExp(GITHUB_APP) }).click();

    await page.getByRole("button", { name: "Show Repos" }).click();
    const repos = page.getByRole("dialog", { name: "Repositories" });
    await repos
        .getByRole("row", { name: /hivepaas-test\/go-app\b(?!-)/ })
        .getByRole("button", { name: "Select" })
        .click();
    await expect(page.getByRole("group", { name: /^Git Repository/ }).getByRole("textbox")).toHaveValue(REPO);

    await page.getByRole("button", { name: "Show Branches" }).click();
    const branches = page.getByRole("dialog", { name: "Branches" });
    // Listed from GitHub as the dialog opens: room for it.
    await expect(branches.getByRole("row", { name: /tiendc-patch-4/ })).toBeVisible({ timeout: 30_000 });
    await branches
        .getByRole("row", { name: /\bmain\b/ })
        .getByRole("button")
        .click();
    await expect(page.getByRole("group", { name: /^Branch/ }).getByRole("textbox")).toHaveValue("refs/heads/main");
    await page.getByRole("button", { name: "Deploy", exact: true }).click();

    await deployed(api, app);
    await expectLogs(page, app, "Server is running");
    const domain = domainFor("github-public");
    await exposeApp(api, app, domain, { port: 8080 });
    await visit(page, `https://${domain}/`);
    await expect(page.locator("body")).toHaveText("Hello");
});

// A private repository is not found without credentials, as GitHub has it for
// a stranger; with the GitHub App's, or the access token's, it is built. What is
// built is a file of it, printed: a Go build each time would only add load.
test("a private GitHub repository is built with the GitHub App's credentials, or a token's, and not without", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "github-private");
    const refused = await saveRepoSource(api, app, { repoURL: PRIVATE_REPO, branch: "main" });
    expect(refused.ok(), "the private repository without credentials").toBe(false);
    expect(((await refused.json()) as { code: string }).code).toBe("ERR_REPO_NOT_FOUND");

    const content = [
        "FROM busybox:1.37",
        "COPY Makefile /Makefile",
        `CMD ["sh", "-c", "echo private-makefile-of-$(grep -c . /Makefile)-lines-by-$CREDENTIALS; exec sleep 3600"]`,
    ].join("\n");
    for (const [credentials, name] of [
        [GITHUB_APP_ID, "app"],
        [GITHUB_TOKEN_ID, "token"],
    ] as const) {
        const res = await saveRepoSource(api, app, {
            repoURL: PRIVATE_REPO,
            branch: "main",
            credentials,
            dockerfile: { source: "manual", path: "Dockerfile", content: content.replace("$CREDENTIALS", name) },
        });
        expect(res.ok(), `building with the ${name}: ${res.status()} ${await res.text()}`).toBe(true);
        const { deploymentId } = ((await res.json()) as { data: { deploymentId: string } }).data;
        await deployed(api, app, deploymentId);
        await expectLogs(page, app, new RegExp(`private-makefile-of-[1-9]\\d*-lines-by-${name}`));
    }
});

// The pull requests of a repository are listed through the GitHub App, each
// with its branch: what a preview is made from.
test("a GitHub repository's pull requests are listed through the GitHub App", async ({ api, cleanup }) => {
    const project = await createProject(api, e2eName("github-pulls"));
    cleanup(() => deleteProject(api, project.id));
    const res = await api.get(`projects/${project.id}/git-credentials/${GITHUB_APP_ID}/repository/pull-requests`, {
        params: { repo: "go-app" },
    });
    expect(res.ok(), `listing the pull requests: ${res.status()} ${await res.text()}`).toBe(true);
    const pulls = ((await res.json()) as { data: { number: number; branch: string; state: string }[] }).data;
    expect(pulls).toEqual(expect.arrayContaining([expect.objectContaining({ number: 4, branch: "tiendc-patch-4" })]));
});

// What a build of go-app finds of its submodule and of a file stored with Git
// LFS: the submodule's files, counted, and the file's first bytes - "PK", a zip,
// once fetched; "ve" of "version https://git-lfs...", the pointer, until then.
const INSPECTING_DOCKERFILE = [
    "FROM busybox:1.37",
    "COPY . /src",
    `CMD ["sh", "-c", "echo found-sub-$(ls -A /src/submodules/mod1 | wc -l)-lfs-$(head -c 2 /src/lfs/go-app.zip); exec sleep 3600"]`,
].join("\n");

// The repository options decide what a build takes of a repository beside its
// own files: its submodules, and the files Git LFS keeps elsewhere. Off, the
// submodule is an empty directory and the file its pointer; on, both are there.
test("a GitHub repository's submodules and LFS files come with its build when asked", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "github-options");
    for (const [repoOptions, found] of [
        [{ gitSubmodulesEnabled: false, gitLfsEnabled: false }, /found-sub-0-lfs-ve/],
        [{ gitSubmodulesEnabled: true, gitLfsEnabled: true }, /found-sub-[1-9]\d*-lfs-PK/],
    ] as const) {
        const res = await saveRepoSource(api, app, {
            repoURL: REPO,
            branch: "main",
            credentials: GITHUB_APP_ID,
            dockerfile: { source: "manual", path: "Dockerfile", content: INSPECTING_DOCKERFILE },
            autoDeploy: false,
            repoOptions,
        });
        expect(res.ok(), `building with ${JSON.stringify(repoOptions)}: ${res.status()} ${await res.text()}`).toBe(
            true,
        );
        const { deploymentId } = ((await res.json()) as { data: { deploymentId: string } }).data;
        await deployed(api, app, deploymentId);
        await expectLogs(page, app, found);
    }
});
