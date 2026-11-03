import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateRegistryAuthStatusDialogOptions } from "../types";

import { useUpdateRegistryAuthStatusDialogState } from "./use-update-registry-auth-status.dialog.state";

function createHook() {
    return function useUpdateRegistryAuthStatusDialog(props: UpdateRegistryAuthStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateRegistryAuthStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    scope: RegistryAuthTableScope,
                    id: string,
                    options: UpdateRegistryAuthStatusDialogOptions = {},
                ) => {
                    actions.open(scope, id, {
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

export const useUpdateRegistryAuthStatusDialog = createHook();
