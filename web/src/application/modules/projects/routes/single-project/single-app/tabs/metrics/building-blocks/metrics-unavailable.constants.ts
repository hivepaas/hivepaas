import type { AppHttpMetricsReason, AppLogHistoryReason } from "~/projects/api/services";

/** Why an app's requests cannot be counted, besides the logs' own reasons, as a person reads it. */
export const HTTP_UNAVAILABLE_TEXT: Record<Exclude<AppHttpMetricsReason, AppLogHistoryReason>, string> = {
    "not-exposed":
        "The app has no domain. Its requests are counted where they enter, at Traefik, and none of its go through it.",
    "access-log-off":
        "Traefik's access log is off. An administrator turns it on in System → Traefik → Config Options, with Access Log; Traefik restarts briefly.",
    "access-log-not-json":
        "Traefik's access log is written in an older form. An administrator saves System → Traefik → Config Options once, with Access Log on; Traefik restarts briefly.",
    "access-log-unlabelled":
        "Traefik's log lines do not carry its identity yet. An administrator saves System → Traefik → Config Options once, with Access Log on; Traefik restarts briefly.",
};

/** Why an app's CPU and memory cannot be read when the agent does not mark its rows. */
export const AGENT_UNLABELLED_TEXT =
    "The HivePaaS agent does not mark its lines yet. It does from its next update, when HivePaaS is updated.";
