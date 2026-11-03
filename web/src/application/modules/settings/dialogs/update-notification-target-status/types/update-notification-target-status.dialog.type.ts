import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

export interface UpdateNotificationTargetStatusDialogState {
    state: { mode: "open"; scope: NotificationTargetTableScope; id: string } | { mode: "closed" };
}

export interface UpdateNotificationTargetStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
