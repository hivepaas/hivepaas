import type { SSHKeyTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSSHKeyDialogOptions } from "../types";

import { useCreateOrEditSSHKeyDialogState } from "./use-create-or-edit-ssh-key.dialog.state";

function createHook() {
    return function useCreateOrEditSSHKeyDialog(props: CreateOrEditSSHKeyDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditSSHKeyDialogState();

        return {
            state,
            actions: {
                open: (scope: SSHKeyTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: SSHKeyTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditSSHKeyDialog = createHook();
