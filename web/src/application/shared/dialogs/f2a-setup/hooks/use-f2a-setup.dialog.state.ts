import { create } from "zustand";

import { type F2aSetupDialogOptions, type F2aSetupDialogState } from "../types";

type State = F2aSetupDialogState & F2aSetupDialogOptions;

interface Actions {
    open: (options: F2aSetupDialogOptions) => void;
    openChange: (options: F2aSetupDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useF2aSetupDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: options => {
        set({
            state: {
                mode: "open",
            },
            ...options,
        });
    },

    openChange: options => {
        set({
            state: {
                mode: "change",
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
