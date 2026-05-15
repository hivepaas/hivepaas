import { create } from "zustand";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditRegistryAuthDialogOptions, CreateOrEditRegistryAuthDialogState } from "../types";

type State = CreateOrEditRegistryAuthDialogState & CreateOrEditRegistryAuthDialogOptions;

interface Actions {
    open: (scope: RegistryAuthTableScope, options?: CreateOrEditRegistryAuthDialogOptions) => void;
    openEdit: (scope: RegistryAuthTableScope, id: string, options?: CreateOrEditRegistryAuthDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditRegistryAuthDialogState = create<State & Actions>()(set => ({
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
