import type { AppStorageMount } from "~/projects/domain";

export interface StorageMountDialogState {
    state:
        | {
              mode: "open";
              projectKey?: string;
              appKey?: string;
          }
        | {
              mode: "edit";
              mount: AppStorageMount & { _id: string };
              projectKey?: string;
              appKey?: string;
          }
        | {
              mode: "closed";
          };
}

export interface StorageMountDialogOptions {
    projectKey?: string;
    appKey?: string;
    props?: {
        onClose?: () => void;
        onSubmit?: (mount: AppStorageMount) => Promise<void>;
        onError?: (error: Error) => void;
        readOnly?: boolean;
    };
}
