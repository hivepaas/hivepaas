import type { AppSecret } from "~/projects/domain";

import type { CreateOrEditAppSecretDialogOptions } from "../types";

import { useCreateOrEditAppSecretDialogState } from "./use-create-or-edit-app-secret.dialog.state";

function createHook() {
    return function useCreateOrEditAppSecretDialog(props: CreateOrEditAppSecretDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useCreateOrEditAppSecretDialogState();

        return {
            state,
            actions: {
                open: (projectId: string, appId: string) => {
                    actions.open(projectId, appId, { props });
                },
                openEdit: (projectId: string, appId: string, secret: AppSecret) => {
                    actions.openEdit(projectId, appId, secret, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditAppSecretDialog = createHook();
