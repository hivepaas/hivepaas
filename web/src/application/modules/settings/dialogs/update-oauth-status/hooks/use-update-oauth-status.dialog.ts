import type { UpdateOAuthStatusDialogOptions } from "../types";

import { useUpdateOAuthStatusDialogState } from "./use-update-oauth-status.dialog.state";

function createHook() {
    return function useUpdateOAuthStatusDialog(props: UpdateOAuthStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateOAuthStatusDialogState();

        return {
            state,
            actions: {
                open: (id: string) => {
                    actions.open(id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateOAuthStatusDialog = createHook();
