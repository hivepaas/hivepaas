import type { AppStorageMount, ProjectStorageSettings } from "~/projects/domain";

export interface StorageMountDialogState {
    state:
        | {
              mode: "open";
              projectRules?: ProjectStorageSettings;
          }
        | {
              mode: "edit";
              mount: AppStorageMount & { _id: string };
              projectRules?: ProjectStorageSettings;
          }
        | {
              mode: "closed";
          };
}

export interface StorageMountDialogOptions {
    props?: {
        onClose?: () => void;
        onSubmit?: (mount: AppStorageMount) => void;
        onError?: (error: Error) => void;
    };
}
