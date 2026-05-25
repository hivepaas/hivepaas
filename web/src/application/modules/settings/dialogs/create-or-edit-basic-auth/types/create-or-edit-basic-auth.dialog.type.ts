import type { BasicAuthTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditBasicAuthDialogState {
    state:
        | {
              mode: "open";
              scope: BasicAuthTableScope;
          }
        | {
              mode: "edit";
              scope: BasicAuthTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditBasicAuthDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
