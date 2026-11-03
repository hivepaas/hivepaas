import { create } from "zustand";
import { type UserBase } from "~/user-management/domain";

import {
    type ResetUserPasswordDialogOptions,
    type ResetUserPasswordDialogState,
} from "../types/reset-user-password.dialog.type";

type State = ResetUserPasswordDialogState & ResetUserPasswordDialogOptions;

interface Actions {
    open: (user: UserBase, options: ResetUserPasswordDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useResetUserPasswordDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        user: null,
    },
    props: {},

    open: (user, options) => {
        set({
            state: {
                mode: "open",
                user,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                user: null,
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
                    user: null,
                },
                props: {},
            };
        });
    },
}));
