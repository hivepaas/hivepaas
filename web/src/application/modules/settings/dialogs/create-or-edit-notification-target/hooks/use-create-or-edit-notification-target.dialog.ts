import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditNotificationTargetDialogOptions } from "../types";

import { useCreateOrEditNotificationTargetDialogState } from "./use-create-or-edit-notification-target.dialog.state";

function createHook() {
    return function useCreateOrEditNotificationTargetDialog(
        props: CreateOrEditNotificationTargetDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useCreateOrEditNotificationTargetDialogState();

        return {
            state,
            actions: {
                open: (scope: NotificationTargetTableScope) => {
                    actions.open(scope, { props });
                },
                openEdit: (scope: NotificationTargetTableScope, id: string) => {
                    actions.openEdit(scope, id, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditNotificationTargetDialog = createHook();
