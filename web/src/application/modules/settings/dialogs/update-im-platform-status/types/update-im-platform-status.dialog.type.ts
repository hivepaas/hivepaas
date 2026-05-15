import type { ImPlatformTableScope } from "~/settings/module-shared/components";

export interface UpdateImPlatformStatusDialogState {
    state:
        | {
              mode: "open";
              scope: ImPlatformTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateImPlatformStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
