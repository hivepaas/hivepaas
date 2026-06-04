import type { CreateFeedbackDialogOptions } from "../types";

import { useCreateFeedbackDialogState } from "./use-create-feedback.dialog.state";

function createHook() {
    return function useCreateFeedbackDialog(props: CreateFeedbackDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateFeedbackDialogState();

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

export const useCreateFeedbackDialog = createHook();
