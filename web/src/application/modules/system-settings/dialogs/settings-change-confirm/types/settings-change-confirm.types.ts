import type { SettingsPendingChange } from "~/system-settings/domain";

/**
 * Which settings are on trial.
 *
 * The dialog is the same either way, but the endpoints are not, and neither is
 * what the change disturbs: "routing" only rewrites labels, while "service"
 * restarts traefik and takes every route down with it for a moment.
 */
export type SettingsChangeKind = "routing" | "service";

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
};

export type SettingsChangeConfirmDialogOptions = {
    props?: {
        onResolved?: (outcome: SettingsChangeOutcome, changeId: string) => void;
    };
};
