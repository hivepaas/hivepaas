import { create } from "zustand";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { UpdateNetworkStatusDialogOptions, UpdateNetworkStatusDialogState } from "../types";

type State = UpdateNetworkStatusDialogState & UpdateNetworkStatusDialogOptions;

interface Actions {
    open: (scope: NetworkManagementScope, id: string, options?: UpdateNetworkStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateNetworkStatusDialogState = create<State & Actions>()(set => ({
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
