import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditAcmeDnsProviderDialogOptions } from "../types";

import { useCreateOrEditAcmeDnsProviderDialogState } from "./use-create-or-edit-acme-dns-provider.dialog.state";

function createHook() {
    return function useCreateOrEditAcmeDnsProviderDialog(
        props: CreateOrEditAcmeDnsProviderDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useCreateOrEditAcmeDnsProviderDialogState();

        return {
            state,
            actions: {
                open: (scope: AcmeDnsProviderTableScope, options: CreateOrEditAcmeDnsProviderDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: AcmeDnsProviderTableScope,
                    id: string,
                    options: CreateOrEditAcmeDnsProviderDialogOptions = {},
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

export const useCreateOrEditAcmeDnsProviderDialog = createHook();
