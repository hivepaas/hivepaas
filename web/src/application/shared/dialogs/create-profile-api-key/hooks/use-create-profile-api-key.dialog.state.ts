import { create } from "zustand";

import { type CreateProfileApiKeyDialogOptions, type CreateProfileApiKeyDialogState } from "../types";

type State = CreateProfileApiKeyDialogState & CreateProfileApiKeyDialogOptions;

interface Actions {
    open: (options?: CreateProfileApiKeyDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateProfileApiKeyDialogState = create<State & Actions>()(set => ({
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

