import type { ImportFilesToContainerDialogOptions } from "../types";

import { useImportFilesToContainerDialogState } from "./use-import-files-to-container.dialog.state";

function createHook() {
    return function useImportFilesToContainerDialog(props: ImportFilesToContainerDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useImportFilesToContainerDialogState();

        return {
            state,
            actions: {
                open: (
                    projectId: string,
                    env: string,
                    appId: string,
                    nodeId: string,
                    containerId: string,
                ) => {
                    actions.open(projectId, env, appId, nodeId, containerId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useImportFilesToContainerDialog = createHook();
