import type { GithubAppTableScope } from "~/settings/module-shared/components";

export interface ProvisionGithubAppDialogState {
    state:
        | {
              mode: "open";
              scope: GithubAppTableScope;
          }
        | {
              mode: "closed";
          };
}

export interface ProvisionGithubAppDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
