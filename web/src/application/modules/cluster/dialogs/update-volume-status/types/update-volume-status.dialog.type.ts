import type { VolumeManagementScope } from "~/cluster/module-shared/types";

export interface UpdateVolumeStatusDialogState {
    state:
        | {
              mode: "open";
              scope: VolumeManagementScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateVolumeStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
