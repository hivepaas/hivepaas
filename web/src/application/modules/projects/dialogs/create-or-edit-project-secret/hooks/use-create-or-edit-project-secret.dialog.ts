import type { AppSecret, ProjectSecret } from "~/projects/domain";

import type { CreateOrEditProjectSecretDialogOptions } from "../types";

import { useCreateOrEditProjectSecretDialogState } from "./use-create-or-edit-project-secret.dialog.state";

function createHook() {
    return function useCreateOrEditProjectSecretDialog(props: CreateOrEditProjectSecretDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditProjectSecretDialogState();

        return {
            state,
            actions: {
                open: (projectId: string) => {
                    actions.open(projectId, "project", undefined, { props });
                },
                openEdit: (projectId: string, secret: ProjectSecret) => {
                    actions.openEdit(projectId, "project", secret, undefined, { props });
                },
                openForApp: (projectId: string, appId: string) => {
                    actions.open(projectId, "app", appId, { props });
                },
                openEditForApp: (projectId: string, appId: string, secret: AppSecret) => {
                    actions.openEdit(projectId, "app", secret, appId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditProjectSecretDialog = createHook();
