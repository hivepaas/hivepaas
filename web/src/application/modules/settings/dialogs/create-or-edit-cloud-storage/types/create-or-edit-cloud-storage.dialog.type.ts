import type { CloudStorageTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditCloudStorageDialogState {
    state:
        | { mode: "open"; scope: CloudStorageTableScope }
        | { mode: "edit"; scope: CloudStorageTableScope; id: string }
        | { mode: "closed" };
}

export interface CreateOrEditCloudStorageDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
