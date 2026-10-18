import type { ManagedRestoreDialogOptions, ManagedRestorePreset, ManagedRestoreSource } from "../types";

import { useManagedRestoreDialogState } from "./use-managed-restore.dialog.state";

function createHook() {
    return function useManagedRestoreDialog(props: ManagedRestoreDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useManagedRestoreDialogState();

        return {
            state,
            actions: {
                open: (source: ManagedRestoreSource, preset?: ManagedRestorePreset) => {
                    actions.open(source, preset, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useManagedRestoreDialog = createHook();
