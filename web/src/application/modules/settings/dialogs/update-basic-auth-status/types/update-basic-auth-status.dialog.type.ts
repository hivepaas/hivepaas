import type { SettingBasicAuth } from "~/settings/domain";
import type { BasicAuthTableScope } from "~/settings/module-shared/components";

export interface UpdateBasicAuthStatusDialogState {
    state:
        | {
              mode: "open";
              scope: BasicAuthTableScope;
              basicAuth: SettingBasicAuth;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateBasicAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
