import { create } from "zustand";

import { type ChangePasswordDialogOptions, type ChangePasswordDialogState } from "../types";

type State = ChangePasswordDialogState & ChangePasswordDialogOptions;

interface Actions {
    open: (options?: ChangePasswordDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useChangePasswordDialogState = create<State & Actions>()(set => ({
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
