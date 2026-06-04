import type { NetworkManagementScope } from "~/cluster/module-shared/types";

export interface CreateNetworkDialogState {
    state:
        | {
              mode: "open";
              scope: NetworkManagementScope;
          }
        | {
              mode: "closed";
              scope: null;
          };
}

export interface CreateNetworkDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
