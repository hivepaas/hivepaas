import type { AppLogHistoryReason } from "~/projects/api/services";

import { LOG_HISTORY_UNAVAILABLE_TEXT } from "../../../tabs/logs/building-blocks/log-history-unavailable.constants";
import {
    AGENT_UNLABELLED_TEXT,
    HTTP_UNAVAILABLE_TEXT,
} from "../../../tabs/metrics/building-blocks/metrics-unavailable.constants";

/** The scale-in delays offered: 1 minute to 1 hour, as the API takes them. */
export const SCALE_IN_DELAYS = [
    { value: "1m", label: "1 minute" },
    { value: "2m", label: "2 minutes" },
    { value: "5m", label: "5 minutes" },
    { value: "10m", label: "10 minutes" },
    { value: "15m", label: "15 minutes" },
    { value: "30m", label: "30 minutes" },
    { value: "1h", label: "1 hour" },
];

/** The reasons an administrator answers in System → Logging. */
export const LOGGING_REASONS = ["disabled", "apps-not-collected"];

/** Why autoscale cannot scale the app at all, rather than why it cannot read what it scales on. */
export const REFUSED_REASONS = ["not-replicated", "host-ports"];

/** Why a signal - a function's calls, an app's requests or CPU - cannot be read, as a person reads it. */
export function signalText(reason: string): string {
    if (reason in HTTP_UNAVAILABLE_TEXT) {
        return HTTP_UNAVAILABLE_TEXT[reason as keyof typeof HTTP_UNAVAILABLE_TEXT];
    }
    if (reason in LOG_HISTORY_UNAVAILABLE_TEXT) {
        return LOG_HISTORY_UNAVAILABLE_TEXT[reason as AppLogHistoryReason];
    }
    switch (reason) {
        case "agent-unlabelled":
            return AGENT_UNLABELLED_TEXT;
        case "no-cpu-limit":
            return "The app has neither a CPU limit nor a CPU reservation to measure its CPU by. Set one in its Resources.";
        default:
            return "It cannot be read now.";
    }
}

/** Why autoscale cannot act, as a person reads it. */
export function pausedText(reason: string, isFunction: boolean): string {
    switch (reason) {
        case "not-replicated":
            return "Autoscale scales an app that runs a set number of instances: set its Service Mode to Replicated.";
        case "host-ports":
            return "The app publishes a port in host mode, so it runs one replica a node at most. Publish it through the routing mesh, in its Network settings, to autoscale it.";
    }
    if (isFunction) {
        return `Autoscale reads the function's calls from its stored logs, which cannot be read. ${signalText(reason)}`;
    }
    return `Autoscale cannot read what the app scales on. ${signalText(reason)}`;
}

/** The replicas the app runs now. A stopped one has none, and autoscale does not start it. */
export function replicasText(replicas: number): string {
    if (replicas === 0) {
        return "Stopped: autoscale does not start it";
    }
    return replicas === 1 ? "1 replica now" : `${replicas} replicas now`;
}
