import type { KeyAuthTableScope } from "~/settings/module-shared/components";

export interface UpdateKeyAuthStatusDialogState {
    state:
        | {
              mode: "open";
              scope: KeyAuthTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateKeyAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
