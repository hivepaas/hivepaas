import type { EmailAccountTableScope } from "~/settings/module-shared/components";

export interface UpdateEmailAccountStatusDialogState {
    state:
        | {
              mode: "open";
              scope: EmailAccountTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateEmailAccountStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
