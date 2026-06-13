import type { SslProviderTableScope } from "~/settings/module-shared/components";

import type { UpdateSslProviderStatusDialogOptions } from "../types";

import { useUpdateSslProviderStatusDialogState } from "./use-update-ssl-provider-status.dialog.state";

function createHook() {
    return function useUpdateSslProviderStatusDialog(props: UpdateSslProviderStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateSslProviderStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    scope: SslProviderTableScope,
                    id: string,
                    options: UpdateSslProviderStatusDialogOptions = {},
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

export const useUpdateSslProviderStatusDialog = createHook();
