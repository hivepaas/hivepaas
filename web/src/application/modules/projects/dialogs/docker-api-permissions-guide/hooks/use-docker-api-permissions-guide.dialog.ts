import type { DockerApiPermissionsGuideDialogOptions } from "../types";

import { useDockerApiPermissionsGuideDialogState } from "./use-docker-api-permissions-guide.dialog.state";

function createHook() {
    return function useDockerApiPermissionsGuideDialog(props: DockerApiPermissionsGuideDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useDockerApiPermissionsGuideDialogState();

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

export const useDockerApiPermissionsGuideDialog = createHook();
