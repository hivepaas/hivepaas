import { create } from "zustand";
import type { EmailAccountTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditEmailAccountDialogOptions, CreateOrEditEmailAccountDialogState } from "../types";

type State = CreateOrEditEmailAccountDialogState & CreateOrEditEmailAccountDialogOptions;

interface Actions {
    open: (scope: EmailAccountTableScope, options?: CreateOrEditEmailAccountDialogOptions) => void;
    openEdit: (scope: EmailAccountTableScope, id: string, options?: CreateOrEditEmailAccountDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditEmailAccountDialogState = create<State & Actions>()(set => ({
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
