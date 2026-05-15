import type { BasicAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateBasicAuthStatusDialogOptions } from "../types";

import { useUpdateBasicAuthStatusDialogState } from "./use-update-basic-auth-status.dialog.state";

function createHook() {
    return function useUpdateBasicAuthStatusDialog(props: UpdateBasicAuthStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateBasicAuthStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: BasicAuthTableScope, id: string) => {
                    actions.open(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateBasicAuthStatusDialog = createHook();
