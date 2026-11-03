export interface SetManagerNodesDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "closed";
          };
}

export interface SetManagerNodesDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
