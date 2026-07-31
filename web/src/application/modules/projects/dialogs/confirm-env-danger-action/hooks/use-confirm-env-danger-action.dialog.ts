import type { ConfirmEnvDangerActionDialogOptions, ConfirmEnvDangerActionTarget, EnvDangerAction } from "../types";

import { useConfirmEnvDangerActionDialogState } from "./use-confirm-env-danger-action.dialog.state";

function createHook() {
    return function useConfirmEnvDangerActionDialog(props: ConfirmEnvDangerActionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useConfirmEnvDangerActionDialogState();

        return {
            state,
            actions: {
                open: (action: EnvDangerAction, target: ConfirmEnvDangerActionTarget) => {
                    actions.open(action, target, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useConfirmEnvDangerActionDialog = createHook();
