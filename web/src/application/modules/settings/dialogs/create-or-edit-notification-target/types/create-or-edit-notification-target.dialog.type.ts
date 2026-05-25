import type { NotificationTargetTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditNotificationTargetDialogState {
    state:
        | { mode: "open"; scope: NotificationTargetTableScope }
        | { mode: "edit"; scope: NotificationTargetTableScope; id: string }
        | { mode: "closed" };
}

export interface CreateOrEditNotificationTargetDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
