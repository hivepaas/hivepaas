import type { AppManagedRestorePoint } from "~/projects/domain";

import type { ManagedRestorePointDialogOptions, ManagedRestorePointTarget } from "../types";

import { useManagedRestorePointDialogState } from "./use-managed-restore-point.dialog.state";

function createHook() {
    return function useManagedRestorePointDialog(props: ManagedRestorePointDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useManagedRestorePointDialogState();

        return {
            state,
            actions: {
                openCreate: (target: ManagedRestorePointTarget) => {
                    actions.openCreate(target, { props });
                },
                openEdit: (target: ManagedRestorePointTarget, point: AppManagedRestorePoint) => {
                    actions.openEdit(target, point, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useManagedRestorePointDialog = createHook();
