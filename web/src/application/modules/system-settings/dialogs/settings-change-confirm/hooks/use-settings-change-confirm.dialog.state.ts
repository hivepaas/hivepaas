import { create } from "zustand";
import type { SettingsPendingChange } from "~/system-settings/domain";

import type { SettingsChangeConfirmDialogState, SettingsChangeKind } from "../types";

type State = SettingsChangeConfirmDialogState;

interface Actions {
    open: (kind: SettingsChangeKind, pendingChange: SettingsPendingChange) => void;
    markResolved: (changeId: string) => void;
    close: () => void;
    destroy: () => void;
}

export const useSettingsChangeConfirmDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },
    resolvedChangeIds: new Set<string>(),

    // Idempotent, and refuses a trial that is already over. Two things open this -
    // the update that started the trial, and the module noticing an unfinished one
    // - and they routinely fire within a moment of each other. Reopening would
    // restart the probe and throw away whichever got there first.
    open: (kind, pendingChange) => {
        set(current => {
            if (current.resolvedChangeIds.has(pendingChange.changeId)) {
                return current;
            }
            if (current.state.mode === "open" && current.state.pendingChange.changeId === pendingChange.changeId) {
                return current;
            }

            return {
                state: {
                    mode: "open",
                    kind,
                    pendingChange,
                },
            };
        });
    },

    markResolved: changeId => {
        set(current => {
            const resolvedChangeIds = new Set(current.resolvedChangeIds);
            resolvedChangeIds.add(changeId);
            return { resolvedChangeIds };
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
            },
        });
    },

    // NOTE: unlike the other dialogs in this module, this one is NOT destroyed on
    // navigation. A change on trial keeps running whether or not the operator is
    // still looking at the page that started it, and closing the only thing that
    // can confirm it would leave them with no way to keep their own change.
    destroy: () => {
        set(state => state);
    },
}));
