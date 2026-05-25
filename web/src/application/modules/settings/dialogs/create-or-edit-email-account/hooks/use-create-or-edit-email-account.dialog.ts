import type { EmailAccountTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditEmailAccountDialogOptions } from "../types";

import { useCreateOrEditEmailAccountDialogState } from "./use-create-or-edit-email-account.dialog.state";

function createHook() {
    return function useCreateOrEditEmailAccountDialog(props: CreateOrEditEmailAccountDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditEmailAccountDialogState();

        return {
            state,
            actions: {
                open: (scope: EmailAccountTableScope, options: CreateOrEditEmailAccountDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: EmailAccountTableScope,
                    id: string,
                    options: CreateOrEditEmailAccountDialogOptions = {},
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

export const useCreateOrEditEmailAccountDialog = createHook();
