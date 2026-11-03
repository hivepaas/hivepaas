import type { UpdateAppHealthCheckStatusDialogOptions } from "../types";

import { useUpdateAppHealthCheckStatusDialogState } from "./use-update-app-health-check-status.dialog.state";

function createHook() {
    return function useUpdateAppHealthCheckStatusDialog(props: UpdateAppHealthCheckStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateAppHealthCheckStatusDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string, healthCheckId: string) => {
                    actions.open(projectId, appId, healthCheckId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateAppHealthCheckStatusDialog = createHook();
