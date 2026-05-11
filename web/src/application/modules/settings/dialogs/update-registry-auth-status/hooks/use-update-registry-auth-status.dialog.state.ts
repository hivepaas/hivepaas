import { create } from "zustand";
import type { SettingRegistryAuth } from "~/settings/domain";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateRegistryAuthStatusDialogOptions, UpdateRegistryAuthStatusDialogState } from "../types";

type State = UpdateRegistryAuthStatusDialogState & UpdateRegistryAuthStatusDialogOptions;

interface Actions {
    open: (
        scope: RegistryAuthTableScope,
        registryAuth: SettingRegistryAuth,
        options?: UpdateRegistryAuthStatusDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateRegistryAuthStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, registryAuth, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
                registryAuth,
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
