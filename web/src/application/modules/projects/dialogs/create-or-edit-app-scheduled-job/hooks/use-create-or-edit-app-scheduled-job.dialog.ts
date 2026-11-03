import type { CreateOrEditAppScheduledJobDialogOptions } from "../types";

import { useCreateOrEditAppScheduledJobDialogState } from "./use-create-or-edit-app-scheduled-job.dialog.state";

function createHook() {
    return function useCreateOrEditAppScheduledJobDialog(
        props: CreateOrEditAppScheduledJobDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useCreateOrEditAppScheduledJobDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string) => {
                    actions.open(projectId, appId, { props });
                },
                openEdit: (projectId: string, appId: string, scheduledJobId: string) => {
                    actions.openEdit(projectId, appId, scheduledJobId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAppScheduledJobDialog = createHook();
