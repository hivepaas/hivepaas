import type { SettingRegistryAuth } from "~/settings/domain";
import type { RegistryAuthTableScope } from "~/settings/module-shared/components";

export interface UpdateRegistryAuthStatusDialogState {
    state:
        | {
              mode: "open";
              scope: RegistryAuthTableScope;
              registryAuth: SettingRegistryAuth;
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
    };
}
