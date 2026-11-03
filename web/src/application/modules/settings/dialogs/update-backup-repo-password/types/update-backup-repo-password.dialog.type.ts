import type { BackupRepoTableScope } from "~/settings/module-shared/components";

export interface UpdateBackupRepoPasswordDialogState {
    state: { mode: "open"; scope: BackupRepoTableScope; id: string; updateVer?: number } | { mode: "closed" };
}

export interface UpdateBackupRepoPasswordDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
