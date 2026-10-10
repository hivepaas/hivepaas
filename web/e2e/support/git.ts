import { type APIRequestContext, type APIResponse, expect } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";

import { type App, appPath } from "./api";

// The repositories env/up.sh serves inside dind, where HivePaaS reaches them.
// Their commits are made at fixed dates: their hashes are these on every env.
export const REPOS = {
    // main: two commits, each a Dockerfile printing built-from-commit-<n>;
    // develop at the first; refs/pull/7/head a third, built-from-pull-request.
    shop: "git://127.0.0.1/e2e/shop.git",
    // An index.html and nothing else.
    site: "git://127.0.0.1/e2e/site.git",
    // A Dockerfile printing the build argument GREETING, and the hash of the
    // build secret BUILD_TOKEN.
    args: "git://127.0.0.1/e2e/args.git",
    // A private repository, built-from-private: over HTTPS with a token, over
    // SSH with a key.
    privateHttps: "https://127.0.0.1:8443/cgi-bin/git/e2e/private.git",
    privateSsh: "git@127.0.0.1:e2e/private.git",
};

// How a Git host names the shop repository: its page's address.
export const SHOP_PAGE = "http://127.0.0.1/e2e/shop";

export const SHOP_COMMITS = {
    first: "2010c034663e785fac6e0d81eb55232d5abfcc0b",
    second: "860f1025bb11e0171549d86ac75c002cad5caa56",
    pullRequest: "d2c8dd9fc61246f415e53d0342bf4d6d5e067e83",
};

// The token the private repository's HTTPS server takes.
export const GIT_TOKEN = "e2e-git-token";

// gitSshKey is the private half of the key the repository's SSH server takes,
// which env/up.sh leaves beside its build.
export function gitSshKey(): string {
    return fs.readFileSync(path.join(import.meta.dirname, "..", "env", ".build", "git-ssh-key"), "utf8");
}

export interface RepoSource {
    repoURL: string;
    branch: string;
    commitHash?: string;
    credentials?: string;
    // A Dockerfile of the repository's, or one written out (content) in its
    // place.
    dockerfile?: { source: "manual"; path: string; content?: string } | { source: "auto" };
    // Whether the build takes the repository's submodules, and its Git LFS
    // files; neither when left out.
    repoOptions?: { gitSubmodulesEnabled: boolean; gitLfsEnabled: boolean };
    // Whether a push to the branch deploys the app: on when left out, as the
    // API has it.
    autoDeploy?: boolean;
    // The container's command, instead of the image's.
    command?: string;
}

// saveRepoSource saves the app as built from a repository's branch - at a
// commit of it, with credentials, when given - and answers what the API said;
// saved, the app is deployed.
export async function saveRepoSource(api: APIRequestContext, app: App, source: RepoSource): Promise<APIResponse> {
    const current = await api.get(`${appPath(app)}/deployment-settings`);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    return api.put(`${appPath(app)}/deployment-settings`, {
        data: {
            entrypoint: "",
            command: source.command ?? "",
            workingDir: "",
            preDeploymentCommand: "",
            postDeploymentCommand: "",
            notification: { successUseDefault: true, failureUseDefault: true },
            activeMethod: "repo",
            repoSource: {
                repoType: "git",
                repoURL: source.repoURL,
                repoRef: source.branch,
                commitHash: source.commitHash ?? "",
                credentials: { id: source.credentials ?? "" },
                dockerfile: source.dockerfile ?? { source: "manual", path: "Dockerfile" },
                pushToRegistry: { id: "" },
                autoDeploy: source.autoDeploy ?? true,
                ...(source.repoOptions ? { repoOptions: source.repoOptions } : {}),
            },
            updateVer,
        },
    });
}

// buildFromRepo sets the app to be built from a repository, which deploys it;
// it answers the deployment's id.
export async function buildFromRepo(api: APIRequestContext, app: App, source: RepoSource): Promise<string> {
    const res = await saveRepoSource(api, app, source);
    expect(res.ok(), `building ${app.name} from ${source.repoURL}: ${res.status()} ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: { deploymentId: string } }).data.deploymentId;
}

// createSetting makes a setting of the installation's, open to every project,
// and answers its id.
export async function createSetting(api: APIRequestContext, kind: string, data: object): Promise<string> {
    const res = await api.post(`settings/${kind}`, { data: { inheritable: true, ...data } });
    expect(res.ok(), `creating ${kind}: ${res.status()} ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: { id: string } }).data.id;
}
