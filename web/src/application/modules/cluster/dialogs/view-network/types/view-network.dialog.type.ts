import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

export interface ViewNetworkDialogState {
    state:
        | {
              mode: "open";
              scope: NetworkManagementScope;
              network: ClusterNetwork;
          }
        | {
              mode: "closed";
              scope: null;
              network: null;
          };
}

export interface ViewNetworkDialogOptions {
    props?: {
        onClose?: () => void;
    };
}
