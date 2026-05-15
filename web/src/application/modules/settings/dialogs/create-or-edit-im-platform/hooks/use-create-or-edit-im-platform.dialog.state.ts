import { create } from "zustand";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditImPlatformDialogOptions, CreateOrEditImPlatformDialogState } from "../types";

type State = CreateOrEditImPlatformDialogState & CreateOrEditImPlatformDialogOptions;

interface Actions {
    open: (scope: ImPlatformTableScope, options?: CreateOrEditImPlatformDialogOptions) => void;
    openEdit: (scope: ImPlatformTableScope, id: string, options?: CreateOrEditImPlatformDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditImPlatformDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
            },
            ...options,
        });
    },

    openEdit: (scope, id, options = {}) => {
        set({
            state: {
                mode: "edit",
                scope,
                id,
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
