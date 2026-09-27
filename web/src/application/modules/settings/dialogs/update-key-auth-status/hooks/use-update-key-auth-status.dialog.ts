import type { KeyAuthTableScope } from "~/settings/module-shared/components";

import type { UpdateKeyAuthStatusDialogOptions } from "../types";

import { useUpdateKeyAuthStatusDialogState } from "./use-update-key-auth-status.dialog.state";

function createHook() {
    return function useUpdateKeyAuthStatusDialog(props: UpdateKeyAuthStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateKeyAuthStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: KeyAuthTableScope, id: string, options: UpdateKeyAuthStatusDialogOptions = {}) => {
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

export const useUpdateKeyAuthStatusDialog = createHook();
