import type { AppHealthCheck } from "~/projects/domain";

export interface CreateOrEditAppHealthCheckDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
          }
        | {
              mode: "edit";
              projectId: string;
              appId: string;
              healthCheck: AppHealthCheck;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
          };
}

export interface CreateOrEditAppHealthCheckDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
