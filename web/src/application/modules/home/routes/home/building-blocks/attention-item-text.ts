import { type AttentionItem, AttentionKind, AttentionScope } from "~/home/domain";

import { ROUTE } from "@application/shared/constants";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

export interface AttentionItemText {
    title: string;
    detail: string;
    /** Where the item is looked at and fixed, and what the link says. */
    link: { to: string; label: string } | null;
}

/** The system screen a HivePaaS service of that name belongs to. */
function systemRoute(subject: string): string {
    const name = subject.toLowerCase();
    if (name.includes("registry")) {
        return ROUTE.systemSettings.registry.$route;
    }
    if (name.includes("log") || name.includes("victoria") || name.includes("vlagent")) {
        return ROUTE.systemSettings.logging.$route;
    }
    if (name.includes("traefik")) {
        return ROUTE.systemSettings.traefik.$route;
    }
    return ROUTE.systemSettings.hivepaas.$route;
}

function appLink(item: AttentionItem, screen: "logs" | "instances"): AttentionItemText["link"] {
    if (item.scope === AttentionScope.System) {
        return { to: systemRoute(item.subject), label: "Open" };
    }
    if (!item.project || !item.app || !item.env) {
        return null;
    }
    const app = ROUTE.projects.single.apps.single;
    return screen === "logs"
        ? { to: app.logs.$route(item.project.id, item.env, item.app.id), label: "View logs" }
        : { to: app.instances.$route(item.project.id, item.env, item.app.id), label: "Open app" };
}

function withError(detail: string, error: string): string {
    return error ? `${detail} · ${error}` : detail;
}

/** How an item reads: the server sends what happened, the screen words it. */
export function describeAttentionItem(item: AttentionItem): AttentionItemText {
    switch (item.kind) {
        case AttentionKind.AppRestarting:
            return {
                title: `${item.subject} keeps restarting`,
                detail: withError(
                    `${item.restarts} ${item.restarts === 1 ? "failure" : "failures"} in the last hour`,
                    item.lastError,
                ),
                link: appLink(item, "logs"),
            };
        case AttentionKind.AppNotRunning:
            return {
                title: `${item.subject} is not running`,
                detail: withError(`${item.running} of ${item.desired} running`, item.lastError),
                link: appLink(item, "instances"),
            };
        case AttentionKind.NodeDown:
            return {
                title: `Node ${item.subject} is down`,
                detail: item.nodeState ? `The swarm reports it as ${item.nodeState}` : "The swarm cannot reach it",
                link: { to: ROUTE.cluster.nodes.$route, label: "Open nodes" },
            };
        case AttentionKind.NodeOvercommitted:
            return {
                title: `Memory limits exceed node memory on ${item.subject}`,
                detail:
                    `Apps may use up to ${formatDataSizeCompact(item.memoryLimitsBytes)} ` +
                    `on a ${formatDataSizeCompact(item.memoryTotalBytes)} node; ` +
                    "when they do, one of them is killed",
                link: { to: ROUTE.cluster.nodes.$route, label: "Open nodes" },
            };
        default:
            return { title: item.subject, detail: item.kind, link: null };
    }
}
