import type { GithubAppTableScope } from "~/settings/module-shared/components";

export interface UpdateGithubAppStatusDialogState {
    state:
        | {
              mode: "open";
              scope: GithubAppTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateGithubAppStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
