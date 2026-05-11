import type { SettingRegistryAuth } from "~/settings/domain";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditRegistryAuthDialogState {
    state:
        | {
              mode: "open";
              scope: RegistryAuthTableScope;
          }
        | {
              mode: "edit";
              scope: RegistryAuthTableScope;
              registryAuth: SettingRegistryAuth;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditRegistryAuthDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
