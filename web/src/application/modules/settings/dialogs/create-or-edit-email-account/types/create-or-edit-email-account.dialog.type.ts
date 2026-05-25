import type { EmailAccountTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditEmailAccountDialogState {
    state:
        | {
              mode: "open";
              scope: EmailAccountTableScope;
          }
        | {
              mode: "edit";
              scope: EmailAccountTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditEmailAccountDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
