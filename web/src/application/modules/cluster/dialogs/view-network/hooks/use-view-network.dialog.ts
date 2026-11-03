import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { ViewNetworkDialogOptions } from "../types";

import { useViewNetworkDialogState } from "./use-view-network.dialog.state";

function createHook() {
    return function useViewNetworkDialog(props: ViewNetworkDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useViewNetworkDialogState();

        return {
            state,
            actions: {
                open: (scope: NetworkManagementScope, network: ClusterNetwork) => {
                    actions.open(scope, network, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useViewNetworkDialog = createHook();
