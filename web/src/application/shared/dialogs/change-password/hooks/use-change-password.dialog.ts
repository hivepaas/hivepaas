import { type ChangePasswordDialogOptions } from "../types";

import { useChangePasswordDialogState } from "./use-change-password.dialog.state";

function createHook() {
    return function useChangePasswordDialog(props: ChangePasswordDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useChangePasswordDialogState();

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

export const useChangePasswordDialog = createHook();
