import { create } from "zustand";

import type {
    AppDangerAction,
    ConfirmAppDangerActionDialogOptions,
    ConfirmAppDangerActionDialogState,
    ConfirmAppDangerActionTarget,
} from "../types";

type State = ConfirmAppDangerActionDialogState & ConfirmAppDangerActionDialogOptions;

interface Actions {
    open: (
        action: AppDangerAction,
        target: ConfirmAppDangerActionTarget,
        options?: ConfirmAppDangerActionDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState: ConfirmAppDangerActionDialogState["state"] = {
    mode: "closed",
    action: null,
    target: null,
};

export const useConfirmAppDangerActionDialogState = create<State & Actions>()(set => ({
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
