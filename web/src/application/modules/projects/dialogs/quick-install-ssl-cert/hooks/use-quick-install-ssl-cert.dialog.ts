import type { QuickInstallSslCertDialogOptions } from "../types";

import { useQuickInstallSslCertDialogState } from "./use-quick-install-ssl-cert.dialog.state";

function createHook() {
    return function useQuickInstallSslCertDialog(props: QuickInstallSslCertDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useQuickInstallSslCertDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, domain: string) => {
                    actions.open(projectId, domain, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useQuickInstallSslCertDialog = createHook();
