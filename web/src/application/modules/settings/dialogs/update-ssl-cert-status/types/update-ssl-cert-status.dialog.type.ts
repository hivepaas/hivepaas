import type { SslCertTableScope } from "~/settings/module-shared/components";

export interface UpdateSslCertStatusDialogState {
    state:
        | {
              mode: "open";
              scope: SslCertTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateSslCertStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
