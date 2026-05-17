import type { CloudStorageTableScope } from "~/settings/module-shared/components";

export interface UpdateCloudStorageStatusDialogState {
    state: { mode: "open"; scope: CloudStorageTableScope; id: string } | { mode: "closed" };
}

export interface UpdateCloudStorageStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
