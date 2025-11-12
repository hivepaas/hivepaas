export interface F2aSetupDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "change";
          }
        | {
              mode: "closed";
          };
}

export interface F2aSetupDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
