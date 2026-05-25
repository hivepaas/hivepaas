import type { BasicAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditBasicAuthDialogOptions } from "../types";

import { useCreateOrEditBasicAuthDialogState } from "./use-create-or-edit-basic-auth.dialog.state";

function createHook() {
    return function useCreateOrEditBasicAuthDialog(props: CreateOrEditBasicAuthDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditBasicAuthDialogState();

        return {
            state,
            actions: {
                open: (scope: BasicAuthTableScope, options: CreateOrEditBasicAuthDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: BasicAuthTableScope,
                    id: string,
                    options: CreateOrEditBasicAuthDialogOptions = {},
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

export const useCreateOrEditBasicAuthDialog = createHook();
