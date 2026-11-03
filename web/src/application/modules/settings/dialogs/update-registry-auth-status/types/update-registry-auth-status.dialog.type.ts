import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

export interface UpdateRegistryAuthStatusDialogState {
    state:
        | {
              mode: "open";
              scope: RegistryAuthTableScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateRegistryAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
