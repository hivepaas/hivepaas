import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

import type { UpdateAcmeDnsProviderStatusDialogOptions } from "../types";

import { useUpdateAcmeDnsProviderStatusDialogState } from "./use-update-acme-dns-provider-status.dialog.state";

function createHook() {
    return function useUpdateAcmeDnsProviderStatusDialog(
        props: UpdateAcmeDnsProviderStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateAcmeDnsProviderStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    scope: AcmeDnsProviderTableScope,
                    id: string,
                    options: UpdateAcmeDnsProviderStatusDialogOptions = {},
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

export const useUpdateAcmeDnsProviderStatusDialog = createHook();
