import type { AccessTokenTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditAccessTokenDialogOptions } from "../types";

import { useCreateOrEditAccessTokenDialogState } from "./use-create-or-edit-access-token.dialog.state";

function createHook() {
    return function useCreateOrEditAccessTokenDialog(props: CreateOrEditAccessTokenDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditAccessTokenDialogState();

        return {
            state,
            actions: {
                open: (scope: AccessTokenTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: AccessTokenTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAccessTokenDialog = createHook();
