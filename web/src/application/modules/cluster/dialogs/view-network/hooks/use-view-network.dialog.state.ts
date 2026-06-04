import { create } from "zustand";
import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { ViewNetworkDialogOptions, ViewNetworkDialogState } from "../types";

type State = ViewNetworkDialogState & ViewNetworkDialogOptions;

interface Actions {
    open: (scope: NetworkManagementScope, network: ClusterNetwork, options?: ViewNetworkDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const closedState: ViewNetworkDialogState["state"] = {
    mode: "closed",
    scope: null,
    network: null,
};

export const useViewNetworkDialogState = create<State & Actions>()(set => ({
    state: closedState,
    props: {},
    open: (scope, network, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
                network,
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
