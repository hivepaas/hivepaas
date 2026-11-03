import { type UserBase } from "~/user-management/domain";

import { type ResetUserPasswordDialogOptions } from "../types/reset-user-password.dialog.type";

import { useResetUserPasswordDialogState } from "./use-reset-user-password.dialog.state";

function createHook() {
    return function useResetUserPasswordDialog(props: ResetUserPasswordDialogOptions["props"]) {
        const { state, ...actions } = useResetUserPasswordDialogState();

        return {
            state,
            actions: {
                open: (user: UserBase) => {
                    actions.open(user, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useResetUserPasswordDialog = createHook();
