import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { CreateNetworkDialogOptions } from "../types";

import { useCreateNetworkDialogState } from "./use-create-network.dialog.state";

function createHook() {
    return function useCreateNetworkDialog(props: CreateNetworkDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateNetworkDialogState();

        return {
            state,
            actions: {
                open: (scope: NetworkManagementScope) => {
                    actions.open(scope, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateNetworkDialog = createHook();
