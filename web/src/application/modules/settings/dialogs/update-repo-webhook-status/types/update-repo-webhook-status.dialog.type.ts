import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

export interface UpdateRepoWebhookStatusDialogState {
    state:
        | {
              mode: "open";
              scope: RepoWebhookTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateRepoWebhookStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
