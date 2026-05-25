import type { SslCertTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditSslCertDialogState {
    state:
        | {
              mode: "open";
              scope: SslCertTableScope;
          }
        | {
              mode: "edit";
              scope: SslCertTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditSslCertDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
