import type { SslProviderTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSslProviderDialogOptions } from "../types";

import { useCreateOrEditSslProviderDialogState } from "./use-create-or-edit-ssl-provider.dialog.state";

function createHook() {
    return function useCreateOrEditSslProviderDialog(props: CreateOrEditSslProviderDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditSslProviderDialogState();

        return {
            state,
            actions: {
                open: (scope: SslProviderTableScope, options: CreateOrEditSslProviderDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: SslProviderTableScope,
                    id: string,
                    options: CreateOrEditSslProviderDialogOptions = {},
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

export const useCreateOrEditSslProviderDialog = createHook();
