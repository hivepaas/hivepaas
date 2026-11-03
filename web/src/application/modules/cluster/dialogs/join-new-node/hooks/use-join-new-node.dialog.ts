import { type JoinNewNodeDialogOptions } from "../types";

import { useJoinNewNodeDialogState } from "./use-join-new-node.dialog.state";

function createHook() {
    return function useJoinNewNodeDialog(props: JoinNewNodeDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useJoinNewNodeDialogState();

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

export const useJoinNewNodeDialog = createHook();
