export interface F2aSetupDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "change";
          }
        | {
              mode: "deactivate";
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
        isSetupRequired?: boolean;
    };
}
