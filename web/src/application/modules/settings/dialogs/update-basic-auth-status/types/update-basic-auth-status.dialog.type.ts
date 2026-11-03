import type { BasicAuthTableScope } from "~/settings/module-shared/components";

export interface UpdateBasicAuthStatusDialogState {
    state:
        | {
              mode: "open";
              scope: BasicAuthTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateBasicAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
