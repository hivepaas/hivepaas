import type { EmailAccountTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditEmailAccountDialogOptions } from "../types";

import { useCreateOrEditEmailAccountDialogState } from "./use-create-or-edit-email-account.dialog.state";

function createHook() {
    return function useCreateOrEditEmailAccountDialog(props: CreateOrEditEmailAccountDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditEmailAccountDialogState();

        return {
            state,
            actions: {
                open: (scope: EmailAccountTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: EmailAccountTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditEmailAccountDialog = createHook();
