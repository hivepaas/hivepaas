import type { BackupRepoTableScope } from "~/settings/module-shared/components";

import type { UpdateBackupRepoStatusDialogOptions } from "../types";

import { useUpdateBackupRepoStatusDialogState } from "./use-update-backup-repo-status.dialog.state";

function createHook() {
    return function useUpdateBackupRepoStatusDialog(props: UpdateBackupRepoStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateBackupRepoStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: BackupRepoTableScope, id: string, options: UpdateBackupRepoStatusDialogOptions = {}) => {
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

export const useUpdateBackupRepoStatusDialog = createHook();
