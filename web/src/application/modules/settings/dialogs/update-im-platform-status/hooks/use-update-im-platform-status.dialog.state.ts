import { create } from "zustand";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

import type { UpdateImPlatformStatusDialogOptions, UpdateImPlatformStatusDialogState } from "../types";

type State = UpdateImPlatformStatusDialogState & UpdateImPlatformStatusDialogOptions;

interface Actions {
    open: (scope: ImPlatformTableScope, id: string, options?: UpdateImPlatformStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateImPlatformStatusDialogState = create<State & Actions>()(set => ({
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
