import type { AccessTokenTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditAccessTokenDialogOptions } from "../types";

import { useCreateOrEditAccessTokenDialogState } from "./use-create-or-edit-access-token.dialog.state";

function createHook() {
    return function useCreateOrEditAccessTokenDialog(props: CreateOrEditAccessTokenDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditAccessTokenDialogState();

        return {
            state,
            actions: {
                open: (scope: AccessTokenTableScope, options: CreateOrEditAccessTokenDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: AccessTokenTableScope,
                    id: string,
                    options: CreateOrEditAccessTokenDialogOptions = {},
                ) => {
                    actions.openEdit(scope, id, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAccessTokenDialog = createHook();
