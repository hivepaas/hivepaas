import type { UpdateAppScheduledJobStatusDialogOptions } from "../types";

import { useUpdateAppScheduledJobStatusDialogState } from "./use-update-app-scheduled-job-status.dialog.state";

function createHook() {
    return function useUpdateAppScheduledJobStatusDialog(
        props: UpdateAppScheduledJobStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateAppScheduledJobStatusDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string, scheduledJobId: string) => {
                    actions.open(projectId, appId, scheduledJobId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateAppScheduledJobStatusDialog = createHook();
