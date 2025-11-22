export interface ChangePasswordDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "closed";
          };
}

export interface ChangePasswordDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
