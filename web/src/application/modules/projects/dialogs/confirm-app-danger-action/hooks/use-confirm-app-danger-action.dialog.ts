import type { AppDangerAction, ConfirmAppDangerActionDialogOptions, ConfirmAppDangerActionTarget } from "../types";

import { useConfirmAppDangerActionDialogState } from "./use-confirm-app-danger-action.dialog.state";

function createHook() {
    return function useConfirmAppDangerActionDialog(props: ConfirmAppDangerActionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useConfirmAppDangerActionDialogState();

        return {
            state,
            actions: {
                open: (action: AppDangerAction, target: ConfirmAppDangerActionTarget) => {
                    actions.open(action, target, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useConfirmAppDangerActionDialog = createHook();
