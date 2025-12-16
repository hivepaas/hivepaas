export interface JoinNewNodeDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "closed";
          };
}

export interface JoinNewNodeDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
