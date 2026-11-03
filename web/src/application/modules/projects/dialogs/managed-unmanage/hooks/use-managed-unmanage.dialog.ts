import type { ManagedUnmanageDialogOptions, ManagedUnmanageTarget } from "../types";

import { useManagedUnmanageDialogState } from "./use-managed-unmanage.dialog.state";

function createHook() {
    return function useManagedUnmanageDialog(props: ManagedUnmanageDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useManagedUnmanageDialogState();

        return {
            state,
            actions: {
                open: (target: ManagedUnmanageTarget) => {
                    actions.open(target, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useManagedUnmanageDialog = createHook();
