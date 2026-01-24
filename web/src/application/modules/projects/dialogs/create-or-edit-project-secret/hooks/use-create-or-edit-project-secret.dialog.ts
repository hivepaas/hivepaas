import type { ProjectSecret } from "~/projects/domain";

import type { CreateOrEditProjectSecretDialogOptions } from "../types";

import { useCreateOrEditProjectSecretDialogState } from "./use-create-or-edit-project-secret.dialog.state";

function createHook() {
    return function useCreateOrEditProjectSecretDialog(props: CreateOrEditProjectSecretDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditProjectSecretDialogState();

        return {
            state,
            actions: {
                open: (projectId: string) => {
                    actions.open(projectId, { props });
                },
                openEdit: (projectId: string, secret: ProjectSecret) => {
                    actions.openEdit(projectId, secret, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditProjectSecretDialog = createHook();
