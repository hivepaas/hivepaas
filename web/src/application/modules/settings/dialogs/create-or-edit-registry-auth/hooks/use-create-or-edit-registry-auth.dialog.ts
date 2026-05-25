import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditRegistryAuthDialogOptions } from "../types";

import { useCreateOrEditRegistryAuthDialogState } from "./use-create-or-edit-registry-auth.dialog.state";

function createHook() {
    return function useCreateOrEditRegistryAuthDialog(props: CreateOrEditRegistryAuthDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditRegistryAuthDialogState();

        return {
            state,
            actions: {
                open: (scope: RegistryAuthTableScope, options: CreateOrEditRegistryAuthDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: RegistryAuthTableScope,
                    id: string,
                    options: CreateOrEditRegistryAuthDialogOptions = {},
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

export const useCreateOrEditRegistryAuthDialog = createHook();
