import { create } from "zustand";

import { type SetManagerNodesDialogOptions, type SetManagerNodesDialogState } from "../types";

type State = SetManagerNodesDialogState & SetManagerNodesDialogOptions;

interface Actions {
    open: (options?: SetManagerNodesDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useSetManagerNodesDialogState = create<State & Actions>()(set => ({
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
