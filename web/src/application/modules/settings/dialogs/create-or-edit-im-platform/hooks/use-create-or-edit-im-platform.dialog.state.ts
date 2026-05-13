import { create } from "zustand";
import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditImPlatformDialogOptions, CreateOrEditImPlatformDialogState } from "../types";

type State = CreateOrEditImPlatformDialogState & CreateOrEditImPlatformDialogOptions;

interface Actions {
    open: (scope: ImPlatformTableScope, options?: CreateOrEditImPlatformDialogOptions) => void;
    openEdit: (
        scope: ImPlatformTableScope,
        imPlatform: SettingImService,
        options?: CreateOrEditImPlatformDialogOptions,
    ) => void;
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

    openEdit: (scope, imPlatform, options = {}) => {
        set({
            state: {
                mode: "edit",
                scope,
                imPlatform,
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
