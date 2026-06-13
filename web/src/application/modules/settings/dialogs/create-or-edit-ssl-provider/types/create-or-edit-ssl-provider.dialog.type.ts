import type { SslProviderTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditSslProviderDialogState {
    state:
        | {
              mode: "open";
              scope: SslProviderTableScope;
          }
        | {
              mode: "edit";
              scope: SslProviderTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditSslProviderDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
