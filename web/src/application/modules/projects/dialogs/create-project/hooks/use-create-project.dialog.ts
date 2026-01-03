import { type CreateProjectDialogOptions } from "../types";

import { useCreateProjectDialogState } from "./use-create-project.dialog.state";

function createHook() {
    return function useCreateProjectDialog(props: CreateProjectDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateProjectDialogState();

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

export const useCreateProjectDialog = createHook();

