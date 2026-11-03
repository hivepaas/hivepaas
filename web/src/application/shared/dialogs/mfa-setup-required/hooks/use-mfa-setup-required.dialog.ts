import { type MfaSetupRequiredDialogOptions } from "../types";

import { useMfaSetupRequiredDialogState } from "./use-mfa-setup-required.dialog.state";

function createHook() {
    return function useMfaSetupRequiredDialog(props: MfaSetupRequiredDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useMfaSetupRequiredDialogState();

        return {
            state,
            actions: {
                open: () => {
                    actions.open({ props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useMfaSetupRequiredDialog = createHook();
