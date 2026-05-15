import type { SslCertTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSslCertDialogOptions } from "../types";

import { useCreateOrEditSslCertDialogState } from "./use-create-or-edit-ssl-cert.dialog.state";

function createHook() {
    return function useCreateOrEditSslCertDialog(props: CreateOrEditSslCertDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditSslCertDialogState();

        return {
            state,
            actions: {
                open: (scope: SslCertTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: SslCertTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditSslCertDialog = createHook();
