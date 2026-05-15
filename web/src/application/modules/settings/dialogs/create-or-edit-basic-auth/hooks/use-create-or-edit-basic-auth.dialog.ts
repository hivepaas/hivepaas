import type { BasicAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditBasicAuthDialogOptions } from "../types";

import { useCreateOrEditBasicAuthDialogState } from "./use-create-or-edit-basic-auth.dialog.state";

function createHook() {
    return function useCreateOrEditBasicAuthDialog(props: CreateOrEditBasicAuthDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditBasicAuthDialogState();

        return {
            state,
            actions: {
                open: (scope: BasicAuthTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: BasicAuthTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditBasicAuthDialog = createHook();
