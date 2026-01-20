import { type CreateProjectAppDialogOptions } from "../types";

import { useCreateProjectAppDialogState } from "./use-create-project-app.dialog.state";

function createHook() {
    return function useCreateProjectAppDialog(props: CreateProjectAppDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateProjectAppDialogState();

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

export const useCreateProjectAppDialog = createHook();
