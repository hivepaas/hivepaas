import type { APIRequestContext, Locator, Page } from "@playwright/test";

import {
    type App,
    appPath,
    createApp,
    createJob,
    createProject,
    createVolume,
    created,
    deleteProject,
    deleteVolume,
    deployImage,
    exposeApp,
    findAppNamed,
    runJob,
    setRuntimeEnvVars,
    storageMounts,
} from "../../support/api";
import {
    BUSYBOX,
    appIn,
    appPage,
    deployed,
    expectInstances,
    expectLogs,
    mountVolume,
    redeploy,
    restart,
} from "../../support/apps";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";
import { newHealthCheck } from "../../support/health-checks";
import { slackField as field, slackMessages, slackTarget } from "../../support/notifications";
import { domainFor, visit } from "../../support/routing";

test.describe.configure({ timeout: 300_000 });

test("an app cloned in its environment runs, with its image and variables, and deploys", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "clone");
    const greeting = e2eName("greeting");
    await setRuntimeEnvVars(api, app, [{ key: "E2E_GREETING", value: greeting, isLiteral: false }]);
    await deployImage(api, app, BUSYBOX, "sh -c 'echo \"hello-$E2E_GREETING\"; exec sleep 3600'");
    await deployed(api, app);

    await page.goto(appPage(app, "app-clone"));
    await page.getByRole("group", { name: "Target Name" }).getByRole("textbox").fill("web-clone");
    await page.getByRole("button", { name: "Save and Clone" }).click();
    await expect(page.getByText("App clone started")).toBeVisible();

    // The clone goes with the project, as the app does.
    await expect.poll(() => findAppNamed(api, app, "web-clone"), { timeout: 120_000 }).toBeDefined();
    const clone = (await findAppNamed(api, app, "web-clone"))!;
    // A clone is made from the app's service, not deployed: it lists no
    // deployment, and its log is what says it runs.
    await expectLogs(page, clone, `hello-${greeting}`);

    // Its deployment settings came along: it deploys on its own.
    await redeploy(page, api, clone);
});

// Each of the app's containers counts its start in /data, signs it with its
// host name - the app's key - and says what /data holds: which app's data a
// container reads is in its log.
const COUNTS_ITS_STARTS =
    "sh -c 'mkdir -p /data; n=$(( $(cat /data/starts 2>/dev/null || echo 0) + 1 )); echo $n >/data/starts; " +
    'touch /data/by-$HOSTNAME; echo "starts=$n files=$(ls /data | tr "\\n" ",")"; exec sleep 3600\'';

// appWithData is an app with a volume mounted at /data, run once: its data is
// starts=1 and by-web.
async function appWithData(page: Page, api: APIRequestContext, cleanup: Cleanup, label: string): Promise<App> {
    // Made before the project, to be removed after it: the app mounts it.
    const volumeId = await createVolume(api, e2eName(`${label}-vol`));
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, label);
    await mountVolume(page, app, e2eName(`${label}-vol`), "/data");
    await deployImage(api, app, BUSYBOX, COUNTS_ITS_STARTS);
    await deployed(api, app);
    await expectLogs(page, app, "starts=1 files=by-web,starts,");
    return app;
}

interface CloneOptions {
    // Target Environment, by name: the app's own when not given.
    env?: string;
    status?: "Active" | "Disabled";
    replicas?: number;
    // The clone's domain in place of the app's one.
    domain?: string;
    // Each a switch of the page: left as the page has it - on - when not given.
    routing?: boolean;
    envVars?: boolean;
    secrets?: boolean;
    configFiles?: boolean;
    periodicJobs?: boolean;
    schedJobs?: boolean;
    volumes?: boolean;
    volumeData?: boolean;
    stopSource?: boolean;
    // Post-Clone Commands, on or off; and a command pipe, by name, added to
    // them, run once the clone exists.
    commands?: boolean;
    commandPipe?: string;
    // A notification target, by name, told how the clone ends.
    notify?: string;
}

// section is a block of the Clone page, by its label.
const section = (page: Page, label: string) => page.getByText(label, { exact: true }).locator("xpath=..");

// setSwitch sets one of the page's switches - a block titled so, or a block's
// Enabled - when the test says how.
async function setSwitch(scope: Page | Locator, title: string, on: boolean | undefined): Promise<void> {
    if (on !== undefined) {
        await scope.getByRole("group", { name: title, exact: true }).getByRole("checkbox").setChecked(on);
    }
}

// lastCloneTask is the task of the app's last clone, if it has one.
async function lastCloneTask(
    api: APIRequestContext,
    app: App,
): Promise<{ status: string; lastError: string } | undefined> {
    const res = await api.get(`${appPath(app)}/tasks`);
    expect(res.ok(), `listing tasks: ${res.status()}`).toBe(true);
    const tasks = (
        (await res.json()) as { data: { type: string; status: string; lastError: string; createdAt: string }[] }
    ).data;
    return tasks.filter(t => t.type === "task:app-clone").sort((a, b) => b.createdAt.localeCompare(a.createdAt))[0];
}

// startClone starts a clone of the app from its Clone page, under a name, as
// the options say.
async function startClone(page: Page, app: App, name: string, options: CloneOptions = {}): Promise<void> {
    await page.goto(appPage(app, "app-clone"));
    await page.getByRole("group", { name: "Target Name" }).getByRole("textbox").fill(name);
    if (options.env) {
        await page.getByRole("group", { name: "Target Environment" }).getByRole("combobox").click();
        await page.getByRole("option", { name: options.env }).click();
    }
    if (options.status) {
        await page.getByRole("group", { name: "Target Status" }).getByRole("combobox").click();
        await page.getByRole("option", { name: options.status }).click();
    }
    if (options.replicas !== undefined) {
        await page.getByRole("spinbutton", { name: "Target Replicas" }).fill(String(options.replicas));
    }
    if (options.domain) {
        await page.getByRole("group", { name: "Domain", exact: true }).getByRole("textbox").fill(options.domain);
    }
    await setSwitch(section(page, "Clone Routing Configuration"), "Enabled", options.routing);
    await setSwitch(page, "Clone Env Variables", options.envVars);
    await setSwitch(page, "Clone Secrets", options.secrets);
    await setSwitch(page, "Clone Config Files", options.configFiles);
    await setSwitch(page, "Clone Periodic Jobs", options.periodicJobs);
    await setSwitch(page, "Clone Scheduled Jobs", options.schedJobs);
    await setSwitch(section(page, "Clone Volumes"), "Enabled", options.volumes);
    await setSwitch(page, "Clone Volume Data", options.volumeData);
    await setSwitch(page, "Stop Source App Before Clone", options.stopSource);
    await setSwitch(section(page, "Post-Clone Commands"), "Enabled", options.commands);
    if (options.commandPipe) {
        await setSwitch(section(page, "Post-Clone Commands"), "Enabled", true);
        await page.getByRole("combobox").filter({ hasText: "select command pipe to add" }).click();
        await page.getByRole("option", { name: options.commandPipe }).click();
        await page.getByRole("button", { name: "Add", exact: true }).click();
    }
    if (options.notify) {
        for (const when of ["On Success", "On Failure"]) {
            // Kept from the last clone, it is left: choosing it again unpicks it.
            const picker = page.getByRole("group", { name: when, exact: true }).getByRole("combobox");
            if ((await picker.textContent()) === options.notify) continue;
            await picker.click();
            await page.getByRole("option", { name: options.notify }).click();
            await expect(picker).toHaveText(options.notify);
            await expect(page.getByRole("option")).toHaveCount(0);
        }
    }
    await page.getByRole("button", { name: "Save and Clone" }).click();
    await expect(page.getByText("App clone started")).toBeVisible();
}

// cloneAs clones the app from its Clone page under a name, as the options say,
// and answers the clone once it is listed in its environment.
async function cloneAs(
    page: Page,
    api: APIRequestContext,
    app: App,
    name: string,
    options: CloneOptions = {},
): Promise<App> {
    await startClone(page, app, name, options);
    const near = { ...app, env: options.env ?? app.env };
    // Listed once made; one that failed is never listed, and its task says why.
    await expect
        .poll(
            async () => {
                if (await findAppNamed(api, near, name)) return "made";
                const task = await lastCloneTask(api, app);
                return task?.status === "failed" ? `failed: ${task.lastError}` : "not yet";
            },
            { timeout: 120_000 },
        )
        .toBe("made");
    return (await findAppNamed(api, near, name))!;
}

// The clone's volumes are copies: what the source wrote is in the clone's, and
// what each writes after is its own. Stopped for the copy, the source is
// started again on its own.
test("an app cloned with its volumes' data runs on a copy, and the source, stopped for it, runs again", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-data");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: true, volumeData: true, stopSource: true });

    // The copy holds the source's first start, and the clone's own.
    await expectLogs(page, clone, "starts=2 files=by-web,by-web-clone,starts,");
    // The source was stopped and started again: a second start, on its own data.
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});

// The clone's volumes made without the data: its directories are its own, and
// empty - its container starts on them.
test("an app cloned with its volumes but not their data starts on empty directories of its own", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-empty");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: true, volumeData: false });

    await expectLogs(page, clone, "starts=1 files=by-web-clone,starts,");
    expect((await storageMounts(api, clone)).map(m => m.target)).toEqual(["/data"]);
    await restart(page, app);
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});

// The volumes not cloned, the clone has none: it never reaches the source's
// data, and the source reads nothing of the clone's.
test("an app cloned without its volumes has none, and leaves the source's data alone", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-none");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: false });

    await expectLogs(page, clone, "starts=1 files=by-web-clone,starts,");
    expect(await storageMounts(api, clone), "the clone mounts nothing").toEqual([]);
    await page.goto(appPage(clone, "persistent-storage"));
    await expect(page.getByRole("row", { name: /\/data/ })).toHaveCount(0);

    await restart(page, app);
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});

// The container says where it runs - the app's name and env - and what it was
// given: a variable from a secret, and a config file's content. It serves the
// same where at /.
const SAYS_WHAT_IT_HAS =
    'sh -c \'mkdir -p /www; echo "served-by=$E2E_WHERE" > /www/index.html; ' +
    'echo "where=$E2E_WHERE token=$E2E_TOKEN_VAR conf=$(cat /etc/e2e.conf 2>/dev/null)"; ' +
    "exec httpd -f -p 80 -h /www'";

interface RichApp {
    app: App;
    token: string;
    conf: string;
    domain: string;
}

// richApp is an app with what a clone may take along: a secret, read through a
// variable; a config file, mounted; and a domain.
async function richApp(page: Page, api: APIRequestContext, cleanup: Cleanup, label: string): Promise<RichApp> {
    const app = await appIn(api, cleanup, label);
    const token = e2eName("token");
    const conf = `conf-${e2eName(label)}`;
    const secret = await api.post(`${appPath(app)}/secrets`, {
        data: { key: "E2E_TOKEN", value: token, base64: false },
    });
    expect(secret.ok(), `creating the secret: ${secret.status()} ${await secret.text()}`).toBe(true);
    await setRuntimeEnvVars(api, app, [
        { key: "E2E_TOKEN_VAR", value: "${secrets.E2E_TOKEN}", isLiteral: false },
        { key: "E2E_WHERE", value: "${HIVEPAAS_APP_NAME}-in-${HIVEPAAS_ENV}", isLiteral: false },
    ]);

    await page.goto(appPage(app, "config-files"));
    await page.getByRole("button", { name: "New Config File" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("E2E_CONF");
    await page.getByRole("group", { name: "Value *" }).getByRole("textbox").fill(conf);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /E2E_CONF/ })).toBeVisible();
    await page.goto(appPage(app, "setting-mounts"));
    await page.getByRole("button", { name: "New Setting Mount" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-conf");
    await page.getByRole("combobox", { name: "Mount From" }).click();
    await page.getByRole("option", { name: "Config file" }).click();
    await page.getByRole("group", { name: "Setting *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "E2E_CONF" }).click();
    await page.getByRole("checkbox", { name: "content", exact: true }).check();
    await page.getByRole("textbox", { name: "content path" }).fill("/etc/e2e.conf");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /e2e-conf/ })).toBeVisible();

    await deployImage(api, app, BUSYBOX, SAYS_WHAT_IT_HAS);
    await deployed(api, app);
    const domain = domainFor(label);
    await exposeApp(api, app, domain, { forceHttps: false });
    await expectLogs(page, app, `where=web-in-development token=${token} conf=${conf}`);
    return { app, token, conf, domain };
}

// A clone into another environment runs there, as many times as asked, at a
// domain of its own, with the app's secret and config file; the app answers at
// its own domain still.
test("an app cloned to another environment runs there, with its replicas, its own domain, its secret and file", async ({
    page,
    api,
    cleanup,
}) => {
    const { app, token, conf, domain } = await richApp(page, api, cleanup, "clone-env");
    const cloneDomain = domainFor("clone-env-prod");

    const clone = await cloneAs(page, api, app, "web", { env: "production", replicas: 2, domain: cloneDomain });

    await expectLogs(page, clone, `where=web-in-production token=${token} conf=${conf}`);
    await expectInstances(page, clone, "2/2");
    await visit(page, `http://${cloneDomain}/`);
    await expect(page.locator("body")).toContainText("served-by=web-in-production");
    await visit(page, `http://${domain}/`);
    await expect(page.locator("body")).toContainText("served-by=web-in-development");
});

// Routing, variables, secrets and config files not cloned: the clone has no
// domain, and its container none of what the app was given.
test("an app cloned without its routing, variables, secrets and config files has none of them", async ({
    page,
    api,
    cleanup,
}) => {
    const { app } = await richApp(page, api, cleanup, "clone-bare");

    const clone = await cloneAs(page, api, app, "web-bare", {
        routing: false,
        envVars: false,
        secrets: false,
        configFiles: false,
    });

    await expectLogs(page, clone, "where= token= conf=");
    const routing = (await (await api.get(`${appPath(clone)}/routing-settings`)).json()) as {
        data: { domains?: unknown[] | null };
    };
    expect(routing.data.domains ?? [], "no domain").toEqual([]);
    const secrets = (await (await api.get(`${appPath(clone)}/secrets`)).json()) as { data?: unknown[] | null };
    expect(secrets.data ?? [], "no secret").toEqual([]);
});

// Disabled, the clone is made and runs nothing.
test("an app cloned disabled is made, and runs no instance", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "clone-off");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);

    const clone = await cloneAs(page, api, app, "web-off", { status: "Disabled" });

    await expectInstances(page, clone, "0/0");
    await expect(page.getByRole("heading", { name: "web-off" }).locator("xpath=..")).toContainText("Disabled");
});

interface Named {
    id: string;
    name: string;
}

// jobsOf are the app's jobs of a kind, by name: its health checks, its
// scheduled jobs.
async function jobsOf(api: APIRequestContext, app: App, path: string): Promise<Named[]> {
    const res = await api.get(`${appPath(app)}/${path}`);
    expect(res.ok(), `listing ${path}: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data?: Named[] | null }).data ?? [];
}

// The app's health check and scheduled job come with a clone, the job running
// in the clone's container; turned off, neither does.
test("an app's health check and scheduled job come with its clone, and run there; turned off, neither does", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "clone-jobs");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    await newHealthCheck(page, app, "e2e-check", "http://127.0.0.1:1/");
    await createJob(api, app, "e2e-host", "sh -c 'echo ran-on-$HOSTNAME'");

    const withJobs = await cloneAs(page, api, app, "web-jobs");
    const without = await cloneAs(page, api, app, "web-nojobs", { periodicJobs: false, schedJobs: false });

    expect((await jobsOf(api, withJobs, "periodic-jobs?kind=healthcheck")).map(j => j.name)).toEqual(["e2e-check"]);
    const jobs = await jobsOf(api, withJobs, "sched-jobs");
    expect(jobs.map(j => j.name)).toEqual(["e2e-host"]);
    await expectLogs(page, withJobs, "ready");
    const run = await runJob(api, withJobs, jobs[0]!.id);
    expect(run.status).toBe("done");
    expect(run.log).toContain("ran-on-web-jobs");

    expect(await jobsOf(api, without, "periodic-jobs?kind=healthcheck")).toEqual([]);
    expect(await jobsOf(api, without, "sched-jobs")).toEqual([]);
});

// commandPipe is a command pipe of the project's, for its apps as the form
// makes it - the Clone page lists the env's: the source's command, run in the
// app, writes what the target's, run in the clone, reads. It answers its name.
async function commandPipe(api: APIRequestContext, app: App, source: string, target: string): Promise<string> {
    const template = async (side: string, command: string) =>
        created(api, `projects/${app.projectId}/command-templates`, {
            name: e2eName(`pipe-${side}`),
            inheritable: true,
            kind: "data-ops",
            command,
            script: "",
            workingDir: "",
            envVars: [],
            argGroups: [],
            tty: false,
        });
    const name = e2eName("pipe");
    await created(api, `projects/${app.projectId}/command-pipes`, {
        name,
        inheritable: true,
        sourceCommand: { id: await template("source", source) },
        targetCommand: { id: await template("target", target) },
    });
    return name;
}

// A command pipe run after the clone - a dump of the app read into the clone,
// in small: the source's command reads a file of the app's, the target's
// writes it into the clone's.
test("a command pipe run after the clone carries what the app has into the clone", async ({ page, api, cleanup }) => {
    const volume = e2eName("clone-pipe-vol");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "clone-pipe");
    await mountVolume(page, app, volume, "/data");
    const note = e2eName("note");
    // The app writes the note; its clone waits for what the pipe brings.
    await deployImage(
        api,
        app,
        BUSYBOX,
        `sh -c 'if [ "$HOSTNAME" = web ]; then echo ${note} > /data/note; echo source-ready; ` +
            'else until [ -f /data/piped ]; do sleep 1; done; echo "piped=$(cat /data/piped)"; fi; exec sleep 3600\'',
    );
    await deployed(api, app);
    await expectLogs(page, app, "source-ready");
    // A command is run as it is, not by a shell: one that writes a file asks for
    // one.
    const pipe = await commandPipe(api, app, "cat /data/note", "sh -c 'cat > /data/piped'");

    const clone = await cloneAs(page, api, app, "web-copy", { volumes: true, volumeData: false, commandPipe: pipe });

    await expectLogs(page, clone, `piped=${note}`);
    // Saved with the clone, the pipe is listed once - and so saved once again.
    const saved = await api.get(`${appPath(app)}/clone-settings`);
    const { commandPipes } = ((await saved.json()) as { data: { commandPipes: { name: string }[] } }).data;
    expect(commandPipes.map(p => p.name)).toEqual([pipe]);
});

// A clone tells the target its settings name how it ended: failed, and why - a
// command run after it failed - leaving no app behind; or done, and what it
// made. Its settings tell nobody until they are told to.
test("a clone tells its notification target that it failed, and why, or that it succeeded", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("clone-told"));
    cleanup(() => deleteProject(api, project.id));
    const target = await slackTarget(page, api, project, "clone-told-hook");
    const app = await createApp(api, project, "web");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    const failing = await commandPipe(api, app, "echo hello", "false");

    await page.goto(appPage(app, "app-clone"));
    for (const when of ["On Success", "On Failure"]) {
        const useDefault = page.getByRole("group", { name: `${when} Use Default`, exact: true });
        await expect(useDefault.getByRole("checkbox"), `${when}: nobody, unless asked`).not.toBeChecked();
    }

    const ofClones = async () =>
        (await slackMessages(page, target.hook)).filter(m => m.title.startsWith(`[${project.name}][web] Clone`));
    await startClone(page, app, "web-broken", { commandPipe: failing, notify: target.name });
    await expect
        .poll(async () => (await ofClones()).map(m => m.title), { timeout: 120_000, intervals: [5_000] })
        .toEqual([`[${project.name}][web] Clone failed`]);
    const [failed] = await ofClones();
    expect(field(failed, "Clone")).toBe("web-broken in development");
    expect(field(failed, "Reason")).toContain("The command exited with code 1");
    expect(field(failed, "See clone details")).toContain(`/projects/${project.id}/`);
    expect(await findAppNamed(api, app, "web-broken"), "a failed clone leaves no app").toBeUndefined();

    // The clone's settings are kept: the failing command is left out of this one.
    const clone = await cloneAs(page, api, app, "web-copy", { commands: false, notify: target.name });
    await expect.poll(async () => (await ofClones()).length, { timeout: 60_000, intervals: [5_000] }).toBe(2);
    const done = (await ofClones())[1];
    expect(done?.title).toBe(`[${project.name}][web] Clone succeeded`);
    expect(field(done, "Clone")).toBe("web-copy in development");
    expect(field(done, "Reason"), "nothing failed").toBeUndefined();
    await expectLogs(page, clone, "ready");

    // The target the page was loaded with, picked again, is unpicked: the field
    // says so, and the settings save without it.
    await page.goto(appPage(app, "app-clone"));
    const onSuccess = page.getByRole("group", { name: "On Success", exact: true }).getByRole("combobox");
    const onFailure = page.getByRole("group", { name: "On Failure", exact: true }).getByRole("combobox");
    await expect(onSuccess).toHaveText(target.name);
    await onSuccess.click();
    await page.getByRole("option", { name: target.name }).click();
    await expect(onSuccess).toHaveText("None");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await expect(page.getByText("App clone settings saved")).toBeVisible();
    await page.reload();
    await expect(onFailure).toHaveText(target.name);
    await expect(onSuccess).toHaveText("None");
});
