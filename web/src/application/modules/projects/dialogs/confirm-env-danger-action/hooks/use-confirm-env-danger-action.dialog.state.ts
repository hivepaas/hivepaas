import { create } from "zustand";

import type {
    ConfirmEnvDangerActionDialogOptions,
    ConfirmEnvDangerActionDialogState,
    ConfirmEnvDangerActionTarget,
    EnvDangerAction,
} from "../types";

type State = ConfirmEnvDangerActionDialogState & ConfirmEnvDangerActionDialogOptions;

interface Actions {
    open: (
        action: EnvDangerAction,
        target: ConfirmEnvDangerActionTarget,
        options?: ConfirmEnvDangerActionDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState: ConfirmEnvDangerActionDialogState["state"] = {
    mode: "closed",
    action: null,
    target: null,
};

export const useConfirmEnvDangerActionDialogState = create<State & Actions>()(set => ({
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
