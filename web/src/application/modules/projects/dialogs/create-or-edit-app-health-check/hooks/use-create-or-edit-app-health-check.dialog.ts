import type { AppHealthCheck } from "~/projects/domain";

import type { CreateOrEditAppHealthCheckDialogOptions } from "../types";

import { useCreateOrEditAppHealthCheckDialogState } from "./use-create-or-edit-app-health-check.dialog.state";

function createHook() {
    return function useCreateOrEditAppHealthCheckDialog(props: CreateOrEditAppHealthCheckDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditAppHealthCheckDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string) => {
                    actions.open(projectId, appId, { props });
                },
                openEdit: (projectId: string, appId: string, healthCheck: AppHealthCheck) => {
                    actions.openEdit(projectId, appId, healthCheck, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAppHealthCheckDialog = createHook();
