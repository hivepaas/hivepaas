import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

import type { UpdateNotificationTargetStatusDialogOptions } from "../types";

import { useUpdateNotificationTargetStatusDialogState } from "./use-update-notification-target-status.dialog.state";

function createHook() {
    return function useUpdateNotificationTargetStatusDialog(
        props: UpdateNotificationTargetStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateNotificationTargetStatusDialogState();

        return {
            state,
            actions: {
                open: (
                    scope: NotificationTargetTableScope,
                    id: string,
                    options: UpdateNotificationTargetStatusDialogOptions = {},
                ) => {
                    actions.open(scope, id, {
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

export const useUpdateNotificationTargetStatusDialog = createHook();
