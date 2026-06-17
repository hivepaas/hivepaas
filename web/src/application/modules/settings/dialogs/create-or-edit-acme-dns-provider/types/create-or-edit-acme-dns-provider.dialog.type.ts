import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditAcmeDnsProviderDialogState {
    state:
        | {
              mode: "open";
              scope: AcmeDnsProviderTableScope;
          }
        | {
              mode: "edit";
              scope: AcmeDnsProviderTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditAcmeDnsProviderDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
