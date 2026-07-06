import type { NetworkManagementScope } from "~/cluster/module-shared/types";

export interface UpdateNetworkStatusDialogState {
    state:
        | {
              mode: "open";
              scope: NetworkManagementScope;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateNetworkStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
