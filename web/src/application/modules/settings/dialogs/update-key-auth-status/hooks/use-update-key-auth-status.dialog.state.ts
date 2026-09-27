import { create } from "zustand";
import type { KeyAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateKeyAuthStatusDialogOptions, UpdateKeyAuthStatusDialogState } from "../types";

type State = UpdateKeyAuthStatusDialogState & UpdateKeyAuthStatusDialogOptions;

interface Actions {
    open: (scope: KeyAuthTableScope, id: string, options?: UpdateKeyAuthStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateKeyAuthStatusDialogState = create<State & Actions>()(set => ({
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
