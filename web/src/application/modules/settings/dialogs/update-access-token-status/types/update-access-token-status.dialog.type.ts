import type { AccessTokenTableScope } from "~/settings/module-shared/components";

export interface UpdateAccessTokenStatusDialogState {
    state: { mode: "open"; scope: AccessTokenTableScope; id: string } | { mode: "closed" };
}

export interface UpdateAccessTokenStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
