import type { AppLogHistoryReason } from "~/projects/api/services";

/** Why an app's stored logs cannot be read, as a person reads it. */
export const LOG_HISTORY_UNAVAILABLE_TEXT: Record<AppLogHistoryReason, string> = {
    "disabled": "Stored logs are off. An administrator can turn them on in System → Logging.",
    "apps-not-collected": "App logs are not collected. An administrator can turn them on in System → Logging.",
    "no-query-endpoint": "Logs go to an external backend HivePaaS has no query endpoint for.",
    "driver-unreadable": "This app's log driver cannot be collected. Switch it to json-file in container settings.",
    "identity-missing":
        "This app was created before logging existed. Save its container settings once so its logs can be identified.",
};
