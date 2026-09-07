import { create } from "zustand";

import type { RestartTraefikDialogOptions, RestartTraefikDialogState } from "../types";

type State = RestartTraefikDialogState & RestartTraefikDialogOptions;

interface Actions {
    open: (options?: RestartTraefikDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useRestartTraefikDialogState = create<State & Actions>()(set => ({
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
