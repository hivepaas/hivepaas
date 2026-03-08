import type { AppSecret, ProjectSecret } from "~/projects/domain";

export type CreateOrEditProjectSecretDialogScope = "project" | "app";

export interface CreateOrEditProjectSecretDialogState {
    state:
        | {
              mode: "open";
              scope: CreateOrEditProjectSecretDialogScope;
              projectId: string;
              appId?: string;
          }
        | {
              mode: "edit";
              scope: CreateOrEditProjectSecretDialogScope;
              projectId: string;
              appId?: string;
              secret: ProjectSecret | AppSecret;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
              scope: null;
          };
}

export interface CreateOrEditProjectSecretDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
