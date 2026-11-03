import { type CreateFunctionDialogOptions } from "../types";

import { useCreateFunctionDialogState } from "./use-create-function.dialog.state";

function createHook() {
    return function useCreateFunctionDialog(props: CreateFunctionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateFunctionDialogState();

        return {
            state,
            actions: {
                open: (projectId: string) => {
                    actions.open(projectId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateFunctionDialog = createHook();
