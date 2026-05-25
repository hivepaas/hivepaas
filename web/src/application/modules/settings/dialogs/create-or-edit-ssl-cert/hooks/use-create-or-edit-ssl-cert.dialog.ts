import type { SslCertTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSslCertDialogOptions } from "../types";

import { useCreateOrEditSslCertDialogState } from "./use-create-or-edit-ssl-cert.dialog.state";

function createHook() {
    return function useCreateOrEditSslCertDialog(props: CreateOrEditSslCertDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditSslCertDialogState();

        return {
            state,
            actions: {
                open: (scope: SslCertTableScope, options: CreateOrEditSslCertDialogOptions = {}) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (scope: SslCertTableScope, id: string, options: CreateOrEditSslCertDialogOptions = {}) => {
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

export const useCreateOrEditSslCertDialog = createHook();
