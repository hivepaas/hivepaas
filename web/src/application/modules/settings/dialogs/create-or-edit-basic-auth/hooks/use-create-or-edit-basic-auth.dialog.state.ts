import { create } from "zustand";
import type { BasicAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditBasicAuthDialogOptions, CreateOrEditBasicAuthDialogState } from "../types";

type State = CreateOrEditBasicAuthDialogState & CreateOrEditBasicAuthDialogOptions;

interface Actions {
    open: (scope: BasicAuthTableScope, options?: CreateOrEditBasicAuthDialogOptions) => void;
    openEdit: (scope: BasicAuthTableScope, id: string, options?: CreateOrEditBasicAuthDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditBasicAuthDialogState = create<State & Actions>()(set => ({
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
