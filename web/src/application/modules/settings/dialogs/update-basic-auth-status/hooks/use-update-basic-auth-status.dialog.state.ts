import { create } from "zustand";
import type { BasicAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateBasicAuthStatusDialogOptions, UpdateBasicAuthStatusDialogState } from "../types";

type State = UpdateBasicAuthStatusDialogState & UpdateBasicAuthStatusDialogOptions;

interface Actions {
    open: (scope: BasicAuthTableScope, id: string, options?: UpdateBasicAuthStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateBasicAuthStatusDialogState = create<State & Actions>()(set => ({
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
