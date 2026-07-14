import { type SetManagerNodesDialogOptions } from "../types";

import { useSetManagerNodesDialogState } from "./use-set-manager-nodes.dialog.state";

function createHook() {
    return function useSetManagerNodesDialog(props: SetManagerNodesDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useSetManagerNodesDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useSetManagerNodesDialog = createHook();
