import { create } from "zustand";

import type { ChangeKekDialogOptions, ChangeKekDialogState } from "../types";

type State = ChangeKekDialogState & ChangeKekDialogOptions;

interface Actions {
    open: (options?: ChangeKekDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useChangeKekDialogState = create<State & Actions>()(set => ({
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
