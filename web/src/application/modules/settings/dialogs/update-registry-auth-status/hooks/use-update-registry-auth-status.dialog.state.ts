import { create } from "zustand";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateRegistryAuthStatusDialogOptions, UpdateRegistryAuthStatusDialogState } from "../types";

type State = UpdateRegistryAuthStatusDialogState & UpdateRegistryAuthStatusDialogOptions;

interface Actions {
    open: (scope: RegistryAuthTableScope, id: string, options?: UpdateRegistryAuthStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateRegistryAuthStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, id, options = {}) => {
        set({
            state: {
                mode: "open",
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
