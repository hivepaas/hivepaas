import { create } from "zustand";

import { type MfaSetupRequiredDialogOptions, type MfaSetupRequiredDialogState } from "../types";

type State = MfaSetupRequiredDialogState & MfaSetupRequiredDialogOptions;

interface Actions {
    open: (options?: MfaSetupRequiredDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useMfaSetupRequiredDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (options = {}) => {
        set({
            state: {
                mode: "open",
            },
            ...options,
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

    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: {
                    mode: "closed",
                },
                props: {},
            };
        });
    },
}));
