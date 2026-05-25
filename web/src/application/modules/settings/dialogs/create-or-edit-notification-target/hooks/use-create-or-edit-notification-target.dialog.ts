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
                open: (
                    scope: NotificationTargetTableScope,
                    options: CreateOrEditNotificationTargetDialogOptions = {},
                ) => {
                    actions.open(scope, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                openEdit: (
                    scope: NotificationTargetTableScope,
                    id: string,
                    options: CreateOrEditNotificationTargetDialogOptions = {},
                ) => {
                    actions.openEdit(scope, id, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useCreateOrEditNotificationTargetDialog = createHook();
