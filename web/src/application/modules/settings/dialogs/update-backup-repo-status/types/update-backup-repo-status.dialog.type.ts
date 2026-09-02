import type { BackupRepoTableScope } from "~/settings/module-shared/components";

export interface UpdateBackupRepoStatusDialogState {
    state: { mode: "open"; scope: BackupRepoTableScope; id: string } | { mode: "closed" };
}

export interface UpdateBackupRepoStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
