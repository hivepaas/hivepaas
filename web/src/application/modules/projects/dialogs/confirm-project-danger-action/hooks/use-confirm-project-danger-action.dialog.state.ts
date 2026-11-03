import { create } from "zustand";

import type {
    ConfirmProjectDangerActionDialogOptions,
    ConfirmProjectDangerActionDialogState,
    ConfirmProjectDangerActionTarget,
    ProjectDangerAction,
} from "../types";

type State = ConfirmProjectDangerActionDialogState & ConfirmProjectDangerActionDialogOptions;

interface Actions {
    open: (
        action: ProjectDangerAction,
        target: ConfirmProjectDangerActionTarget,
        options?: ConfirmProjectDangerActionDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState: ConfirmProjectDangerActionDialogState["state"] = {
    mode: "closed",
    action: null,
    target: null,
};

export const useConfirmProjectDangerActionDialogState = create<State & Actions>()(set => ({
    state: closedState,
    props: {},
    open: (action, target, options = {}) => {
        set({
            state: {
                mode: "open",
                action,
                target,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: closedState,
        });
    },
    clear: () => {
        set({
            props: {},
        });
    },
    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: closedState,
                props: {},
            };
        });
    },
}));
