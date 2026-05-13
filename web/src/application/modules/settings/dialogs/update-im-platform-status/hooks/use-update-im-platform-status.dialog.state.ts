import { create } from "zustand";
import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { UpdateImPlatformStatusDialogOptions, UpdateImPlatformStatusDialogState } from "../types";

type State = UpdateImPlatformStatusDialogState & UpdateImPlatformStatusDialogOptions;

interface Actions {
    open: (
        scope: ImPlatformTableScope,
        imPlatform: SettingImService,
        options?: UpdateImPlatformStatusDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateImPlatformStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, imPlatform, options = {}) => {
        set({
            state: {
                mode: "open",
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
