import { create } from "zustand";

import type { RestartHivePaaSDialogOptions, RestartHivePaaSDialogState } from "../types";

type State = RestartHivePaaSDialogState & RestartHivePaaSDialogOptions;

interface Actions {
    open: (options?: RestartHivePaaSDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useRestartHivePaaSDialogState = create<State & Actions>()(set => ({
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
