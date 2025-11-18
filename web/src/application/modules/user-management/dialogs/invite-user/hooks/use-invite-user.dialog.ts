import { type InviteUserDialogOptions } from "../types/invite-user.dialog.type";

import { useInviteUserDialogState } from "./use-invite-user.dialog.state";

function createHook() {
    return function useInviteUserDialog(props: InviteUserDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useInviteUserDialogState();

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

export const useInviteUserDialog = createHook();

