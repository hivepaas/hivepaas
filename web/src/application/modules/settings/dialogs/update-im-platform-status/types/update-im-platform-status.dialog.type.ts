import type { SettingImService } from "~/settings/domain";
import type { ImPlatformTableScope } from "~/settings/module-shared/components";

export interface UpdateImPlatformStatusDialogState {
    state:
        | {
              mode: "open";
              scope: ImPlatformTableScope;
              imPlatform: SettingImService;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateImPlatformStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
