import type { BackupRepoTableScope } from "~/settings/module-shared/components";

import type { UpdateBackupRepoPasswordDialogOptions } from "../types";

import { useUpdateBackupRepoPasswordDialogState } from "./use-update-backup-repo-password.dialog.state";

function createHook() {
    return function useUpdateBackupRepoPasswordDialog(props: UpdateBackupRepoPasswordDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateBackupRepoPasswordDialogState();

        return {
            state,
            actions: {
                open: (
                    scope: BackupRepoTableScope,
                    id: string,
                    updateVer?: number,
                    options: UpdateBackupRepoPasswordDialogOptions = {},
                ) => {
                    actions.open(scope, id, updateVer, {
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

export const useUpdateBackupRepoPasswordDialog = createHook();
