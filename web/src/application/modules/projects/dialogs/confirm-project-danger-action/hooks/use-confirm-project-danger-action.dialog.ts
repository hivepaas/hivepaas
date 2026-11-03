import type {
    ConfirmProjectDangerActionDialogOptions,
    ConfirmProjectDangerActionTarget,
    ProjectDangerAction,
} from "../types";

import { useConfirmProjectDangerActionDialogState } from "./use-confirm-project-danger-action.dialog.state";

function createHook() {
    return function useConfirmProjectDangerActionDialog(props: ConfirmProjectDangerActionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useConfirmProjectDangerActionDialogState();

        return {
            state,
            actions: {
                open: (action: ProjectDangerAction, target: ConfirmProjectDangerActionTarget) => {
                    actions.open(action, target, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useConfirmProjectDangerActionDialog = createHook();
