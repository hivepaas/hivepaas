import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditRepoWebhookDialogState {
    state:
        | {
              mode: "open";
              scope: RepoWebhookTableScope;
          }
        | {
              mode: "edit";
              scope: RepoWebhookTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditRepoWebhookDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
