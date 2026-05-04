import type { AppStorageMount, ProjectStorageSettings } from "~/projects/domain";

export interface StorageMountDialogState {
    state:
        | {
              mode: "open";
              projectRules?: ProjectStorageSettings;
              projectKey?: string;
              appLocalKey?: string;
          }
        | {
              mode: "edit";
              mount: AppStorageMount & { _id: string };
              projectRules?: ProjectStorageSettings;
              projectKey?: string;
              appLocalKey?: string;
          }
        | {
              mode: "closed";
          };
}

export interface StorageMountDialogOptions {
    projectKey?: string;
    appLocalKey?: string;
    props?: {
        onClose?: () => void;
        onSubmit?: (mount: AppStorageMount) => void;
        onError?: (error: Error) => void;
    };
}
