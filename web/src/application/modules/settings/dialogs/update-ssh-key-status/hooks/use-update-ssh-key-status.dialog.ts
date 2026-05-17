import type { SSHKeyTableScope } from "~/settings/module-shared/components";

import type { UpdateSSHKeyStatusDialogOptions } from "../types";

import { useUpdateSSHKeyStatusDialogState } from "./use-update-ssh-key-status.dialog.state";

function createHook() {
    return function useUpdateSSHKeyStatusDialog(props: UpdateSSHKeyStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateSSHKeyStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: SSHKeyTableScope, id: string) => {
                    actions.open(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateSSHKeyStatusDialog = createHook();
