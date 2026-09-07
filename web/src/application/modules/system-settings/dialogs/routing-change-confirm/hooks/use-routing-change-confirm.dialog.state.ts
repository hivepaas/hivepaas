import { create } from "zustand";
import type { HivePaaSRoutingPendingChange } from "~/system-settings/domain";

import type { RoutingChangeConfirmDialogOptions, RoutingChangeConfirmDialogState } from "../types";

type State = RoutingChangeConfirmDialogState & RoutingChangeConfirmDialogOptions;

interface Actions {
    open: (pendingChange: HivePaaSRoutingPendingChange, options?: RoutingChangeConfirmDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useRoutingChangeConfirmDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },
    props: {},

    // Idempotent for the same change. Two things open this dialog - the update
    // that started the trial, and the page noticing an unfinished one - and they
    // routinely fire within a moment of each other. Reopening would restart the
    // probe and throw away whichever of them got there first.
    open: (pendingChange, options = {}) => {
        set(current => {
            if (current.state.mode === "open" && current.state.pendingChange.changeId === pendingChange.changeId) {
                return current;
            }

            return {
                state: {
                    mode: "open",
                    pendingChange,
                },
                ...options,
            };
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
            },
        });
    },

    clear: () => {
        set({
            props: {},
        });
    },

    // NOTE: unlike the other dialogs in this module, this one is NOT destroyed on
    // navigation. A change on trial keeps running whether or not the operator is
    // still looking at the routing page, and closing the only thing that can
    // confirm it would leave them with no way to keep their own change.
    destroy: () => {
        set(state => state);
    },
}));
