import { create } from "zustand";

import { type InviteUserDialogOptions, type InviteUserDialogState } from "../types/invite-user.dialog.type";

type State = InviteUserDialogState & InviteUserDialogOptions;

interface Actions {
    open: (options?: InviteUserDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useInviteUserDialogState = create<State & Actions>()(set => ({
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

