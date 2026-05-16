import type { AppStorageMount } from "~/projects/domain";

export interface StorageMountDialogState {
    state:
        | {
              mode: "open";
              projectKey?: string;
              appLocalKey?: string;
          }
        | {
              mode: "edit";
              mount: AppStorageMount & { _id: string };
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
        onSubmit?: (mount: AppStorageMount) => Promise<void>;
        onError?: (error: Error) => void;
    };
}
