import type { HivePaaSRoutingPendingChange } from "~/system-settings/domain";

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
export type RoutingChangeOutcome = "confirmed" | "reverted" | "expired" | "superseded";

export type RoutingChangeConfirmDialogState = {
    state: { mode: "closed" } | { mode: "open"; pendingChange: HivePaaSRoutingPendingChange };
};

export type RoutingChangeConfirmDialogOptions = {
    props?: {
        onResolved?: (outcome: RoutingChangeOutcome, changeId: string) => void;
    };
};
