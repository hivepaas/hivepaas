import type { EmailAccountTableScope } from "~/settings/module-shared/components";

import type { UpdateEmailAccountStatusDialogOptions } from "../types";

import { useUpdateEmailAccountStatusDialogState } from "./use-update-email-account-status.dialog.state";

function createHook() {
    return function useUpdateEmailAccountStatusDialog(props: UpdateEmailAccountStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateEmailAccountStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: EmailAccountTableScope, id: string) => {
                    actions.open(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateEmailAccountStatusDialog = createHook();
