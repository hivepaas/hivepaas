import { create } from "zustand";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { CreateNetworkDialogOptions, CreateNetworkDialogState } from "../types";

type State = CreateNetworkDialogState & CreateNetworkDialogOptions;

interface Actions {
    open: (scope: NetworkManagementScope, options?: CreateNetworkDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState: CreateNetworkDialogState["state"] = {
    mode: "closed",
    scope: null,
};

export const useCreateNetworkDialogState = create<State & Actions>()(set => ({
    state: closedState,
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
    close: () => {
        set({
            state: closedState,
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
                state: closedState,
                props: {},
            };
        });
    },
}));
