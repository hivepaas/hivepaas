import type { ProjectSecret } from "~/projects/domain";

export interface CreateOrEditProjectSecretDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
          }
        | {
              mode: "edit";
              projectId: string;
              secret: ProjectSecret;
          }
        | {
              mode: "closed";
              projectId: null;
          };
}

export interface CreateOrEditProjectSecretDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
