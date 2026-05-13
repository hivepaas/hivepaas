import type { SettingSslCert } from "~/settings/domain";
import type { SslCertTableScope } from "~/settings/module-shared/components";

import type { UpdateSslCertStatusDialogOptions } from "../types";

import { useUpdateSslCertStatusDialogState } from "./use-update-ssl-cert-status.dialog.state";

function createHook() {
    return function useUpdateSslCertStatusDialog(props: UpdateSslCertStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateSslCertStatusDialogState();

        return {
            state,
            actions: {
                open: (scope: SslCertTableScope, sslCert: SettingSslCert) => {
                    actions.open(scope, sslCert, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateSslCertStatusDialog = createHook();
