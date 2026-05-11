import type { SettingRegistryAuth } from "~/settings/domain";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditRegistryAuthDialogOptions } from "../types";

import { useCreateOrEditRegistryAuthDialogState } from "./use-create-or-edit-registry-auth.dialog.state";

function createHook() {
    return function useCreateOrEditRegistryAuthDialog(props: CreateOrEditRegistryAuthDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditRegistryAuthDialogState();

        return {
            state,
            actions: {
                open: (scope: RegistryAuthTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: RegistryAuthTableScope, registryAuth: SettingRegistryAuth) => {
                    actions.openEdit(scope, registryAuth, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditRegistryAuthDialog = createHook();
