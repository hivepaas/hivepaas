import type { SettingBasicAuth } from "~/settings/domain";
import type { BasicAuthTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditBasicAuthDialogState {
    state:
        | {
              mode: "open";
              scope: BasicAuthTableScope;
          }
        | {
              mode: "edit";
              scope: BasicAuthTableScope;
              basicAuth: SettingBasicAuth;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditBasicAuthDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
