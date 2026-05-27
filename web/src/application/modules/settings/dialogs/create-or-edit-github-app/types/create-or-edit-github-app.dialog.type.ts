import type { GithubAppTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditGithubAppDialogState {
    state:
        | {
              mode: "open";
              scope: GithubAppTableScope;
          }
        | {
              mode: "edit";
              scope: GithubAppTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditGithubAppDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
