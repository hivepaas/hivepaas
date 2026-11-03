import { create } from "zustand";

import { type JoinNewNodeDialogOptions, type JoinNewNodeDialogState } from "../types";

type State = JoinNewNodeDialogState & JoinNewNodeDialogOptions;

interface Actions {
    open: (options?: JoinNewNodeDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useJoinNewNodeDialogState = create<State & Actions>()(set => ({
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
