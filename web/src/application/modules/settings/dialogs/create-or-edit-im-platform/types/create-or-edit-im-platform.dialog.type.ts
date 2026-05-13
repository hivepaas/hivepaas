import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditImPlatformDialogState {
    state:
        | {
              mode: "open";
              scope: ImPlatformTableScope;
          }
        | {
              mode: "edit";
              scope: ImPlatformTableScope;
              imPlatform: SettingImService;
          }
        | {
              mode: "closed";
          };
}

export interface CreateOrEditImPlatformDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
