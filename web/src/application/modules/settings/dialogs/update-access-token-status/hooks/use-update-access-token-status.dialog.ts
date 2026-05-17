import type { AccessTokenTableScope } from "~/settings/module-shared/components";

import type { UpdateAccessTokenStatusDialogOptions } from "../types";

import { useUpdateAccessTokenStatusDialogState } from "./use-update-access-token-status.dialog.state";

function createHook() {
    return function useUpdateAccessTokenStatusDialog(props: UpdateAccessTokenStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateAccessTokenStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: AccessTokenTableScope, id: string) => {
                    actions.open(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateAccessTokenStatusDialog = createHook();
