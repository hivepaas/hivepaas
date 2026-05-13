import type { SettingSslCert } from "~/settings/domain";
import type { SslCertTableScope } from "~/settings/module-shared/components";

export interface UpdateSslCertStatusDialogState {
    state:
        | {
              mode: "open";
              scope: SslCertTableScope;
              sslCert: SettingSslCert;
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
    };
}
