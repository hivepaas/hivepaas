import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import type { UpdateNetworkStatusDialogOptions } from "../types";

import { useUpdateNetworkStatusDialogState } from "./use-update-network-status.dialog.state";

function createHook() {
    return function useUpdateNetworkStatusDialog(props: UpdateNetworkStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateNetworkStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: NetworkManagementScope, id: string, options: UpdateNetworkStatusDialogOptions = {}) => {
                    actions.open(scope, id, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateNetworkStatusDialog = createHook();
