import type { SslProviderTableScope } from "~/settings/module-shared/components";

export interface UpdateSslProviderStatusDialogState {
    state:
        | {
              mode: "open";
              scope: SslProviderTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateSslProviderStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
