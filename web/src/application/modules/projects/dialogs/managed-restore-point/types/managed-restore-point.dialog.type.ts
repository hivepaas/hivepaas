import type { AppManagedRestorePoint } from "~/projects/domain";

/** The managed database a restore point is made in. */
export interface ManagedRestorePointTarget {
    projectId: string;
    env: string;
    appId: string;
}

export interface ManagedRestorePointDialogState {
    state:
        | { mode: "create"; target: ManagedRestorePointTarget; point: null }
        | { mode: "edit"; target: ManagedRestorePointTarget; point: AppManagedRestorePoint }
        | { mode: "closed"; target: null; point: null };
}

export interface ManagedRestorePointDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
