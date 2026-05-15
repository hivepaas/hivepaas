import type { ImPlatformTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditImPlatformDialogState {
    state:
        | {
              mode: "open";
              scope: ImPlatformTableScope;
          }
        | {
              mode: "edit";
              scope: ImPlatformTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditImPlatformDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
