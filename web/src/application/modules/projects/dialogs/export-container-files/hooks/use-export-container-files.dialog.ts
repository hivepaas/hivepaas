import type { ExportContainerFilesDialogOptions } from "../types";

import { useExportContainerFilesDialogState } from "./use-export-container-files.dialog.state";

function createHook() {
    return function useExportContainerFilesDialog(props: ExportContainerFilesDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useExportContainerFilesDialogState();

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

export const useExportContainerFilesDialog = createHook();
