import type { AccessTokenTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditAccessTokenDialogState {
    state:
        | { mode: "open"; scope: AccessTokenTableScope }
        | { mode: "edit"; scope: AccessTokenTableScope; id: string }
        | { mode: "closed" };
}

export interface CreateOrEditAccessTokenDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
