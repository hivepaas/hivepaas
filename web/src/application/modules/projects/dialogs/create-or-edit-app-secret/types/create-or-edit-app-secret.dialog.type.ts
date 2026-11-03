import type { AppSecret } from "~/projects/domain";

export interface CreateOrEditAppSecretDialogState {
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
              secret: AppSecret;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
          };
}

export interface CreateOrEditAppSecretDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
