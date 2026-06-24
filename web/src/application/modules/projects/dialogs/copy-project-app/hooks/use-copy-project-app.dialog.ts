import { type CopyProjectAppDialogOptions } from "../types";

import { useCopyProjectAppDialogState } from "./use-copy-project-app.dialog.state";

function createHook() {
    return function useCopyProjectAppDialog(props: CopyProjectAppDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCopyProjectAppDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string) => {
                    actions.open(projectId, appId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCopyProjectAppDialog = createHook();
