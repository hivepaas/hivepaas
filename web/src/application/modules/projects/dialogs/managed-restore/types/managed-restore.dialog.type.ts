import type { ManagedRestoreTargetKind } from "~/projects/domain";

/** The managed database restored from: it is only read. */
export interface ManagedRestoreSource {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
}

/** What the dialog opens on: a target already picked, such as a point's "Restore to it". */
export interface ManagedRestorePreset {
    kind: ManagedRestoreTargetKind;
    point?: string;
    backup?: string;
}

export interface ManagedRestoreDialogState {
    state:
        | { mode: "open"; source: ManagedRestoreSource; preset: ManagedRestorePreset | null }
        | { mode: "closed"; source: null; preset: null };
}

export interface ManagedRestoreDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
