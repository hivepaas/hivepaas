import type { AcmeDnsProviderTableScope } from "~/settings/module-shared/components";

export interface UpdateAcmeDnsProviderStatusDialogState {
    state:
        | {
              mode: "open";
              scope: AcmeDnsProviderTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateAcmeDnsProviderStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
