import type { SettingsPendingChange } from "~/system-settings/domain";

/**
 * Which settings are on trial.
 *
 * The dialog is the same for all of them, but the endpoints are not, and neither
 * is what the change disturbs: "routing" only rewrites labels, while "service"
 * and "traefik" both replace Traefik's task and take every route down with it for
 * a moment.
 *
 * "traefik" is the one where the trial is doing work nothing else can. Traefik's
 * healthcheck asks /ping, which answers 200 whether or not a single router was
 * discovered - so a command line that starts cleanly and serves nothing passes
 * every automatic check there is, and only somebody failing to come back reveals
 * it.
 */
export type SettingsChangeKind = "routing" | "service" | "traefik";

/**
 * How the trial ended.
 *
 * "expired" is not the same as "reverted": reverted is something the operator
 * asked for, expired is the deadline passing with nobody vouching for the
 * change. They read very differently to the person who caused them.
 *
 * "superseded" is the honest answer when the server stops reporting the trial and
 * this session did not end it - another session confirmed it, reverted it, or
 * started a change of its own, and there is no way to tell which from here.
 */
export type SettingsChangeOutcome = "confirmed" | "reverted" | "expired" | "superseded";

export type SettingsChangeConfirmDialogState = {
    state: { mode: "closed" } | { mode: "open"; kind: SettingsChangeKind; pendingChange: SettingsPendingChange };

    /**
     * Trials this page has already seen through to an end.
     *
     * It lives in the store rather than in whichever component opened the dialog,
     * because several things open it - the update that started the trial, and the
     * module noticing an unfinished one on any page - and a dedupe that only one
     * of them knows about would let the others reopen a change moments after it
     * was confirmed, from a query result that had not caught up yet.
     */
    resolvedChangeIds: Set<string>;
};
