import type { SSHKeyTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSSHKeyDialogOptions } from "../types";

import { useCreateOrEditSSHKeyDialogState } from "./use-create-or-edit-ssh-key.dialog.state";

function createHook() {
    return function useCreateOrEditSSHKeyDialog(props: CreateOrEditSSHKeyDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditSSHKeyDialogState();

        return {
            state,
            actions: {
                open: (scope: SSHKeyTableScope, options: CreateOrEditSSHKeyDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (scope: SSHKeyTableScope, id: string, options: CreateOrEditSSHKeyDialogOptions = {}) => {
                    actions.openEdit(scope, id, {
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

export const useCreateOrEditSSHKeyDialog = createHook();
