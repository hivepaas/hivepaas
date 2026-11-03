import type { APIRequestContext, Page } from "@playwright/test";

import { type App, type Project, createApp, created, deployImage, exposeApp, setRuntimeEnvVars } from "./api";
import { copyShownLogs, deployed, expectLogs } from "./apps";
import { e2eName, expect } from "./fixtures";
import { domainFor, visit } from "./routing";

// What stands in for Slack: an echo server, which logs each request it gets -
// body and all, on a line of its own with LOG_WITHOUT_NEWLINE.
export const ECHO = "mendhak/http-https-echo:42";

// echoHook is an app of the project's running the echo server, at a domain
// named after the label, over plain HTTP: what calls it - the backend - trusts
// only certificates someone signed, and a .localhost one is signed by no one.
export async function echoHook(
    page: Page,
    api: APIRequestContext,
    project: Pick<Project, "id">,
    label: string,
): Promise<{ hook: App; domain: string }> {
    const hook = await createApp(api, project, "hook");
    await setRuntimeEnvVars(api, hook, [{ key: "LOG_WITHOUT_NEWLINE", value: "true", isLiteral: true }]);
    await deployImage(api, hook, ECHO);
    await deployed(api, hook);
    const domain = domainFor(label);
    await exposeApp(api, hook, domain, { port: 8080, forceHttps: false });
    await visit(page, `http://${domain}/`);
    return { hook, domain };
}

export interface SlackTarget {
    // The echo app that stands in for Slack.
    hook: App;
    // The notification target sending to it: its id, and its name.
    id: string;
    name: string;
}

// slackTarget is a notification target of the project's that sends to an echo
// app standing in for Slack, through the project's Slack: both for the
// project's apps too, as the forms make them.
export async function slackTarget(
    page: Page,
    api: APIRequestContext,
    project: Pick<Project, "id">,
    label: string,
): Promise<SlackTarget> {
    const { hook, domain } = await echoHook(page, api, project, label);
    const platform = await created(api, `projects/${project.id}/im-services`, {
        name: e2eName("slack"),
        inheritable: true,
        kind: "slack",
        slack: { webhook: `http://${domain}/slack` },
    });
    const name = e2eName("to-slack");
    const id = await created(api, `projects/${project.id}/notifications`, {
        name,
        inheritable: true,
        viaSlack: { enabled: true, useDefault: false, webhook: { id: platform } },
    });
    return { hook, id, name };
}

export interface SlackAttachment {
    title: string;
    fields: { title: string; value: string }[];
}

// slackMessages are what the hook was sent at /slack, in the order it came.
export async function slackMessages(page: Page, hook: App): Promise<SlackAttachment[]> {
    // The line the server starts with: the log is there once it is shown.
    await expectLogs(page, hook, "Listening on ports");
    return (await copyShownLogs(page)).split("\n").flatMap(line => {
        try {
            const req = JSON.parse(line) as { path?: string; json?: { attachments?: SlackAttachment[] } };
            return req.path === "/slack" ? (req.json?.attachments ?? []) : [];
        } catch {
            return [];
        }
    });
}

// slackField is the value of a message's field, by its title.
export const slackField = (message: SlackAttachment | undefined, title: string) =>
    message?.fields.find(f => f.title === title)?.value;
