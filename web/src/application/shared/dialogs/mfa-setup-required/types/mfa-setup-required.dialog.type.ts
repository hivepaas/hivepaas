export interface MfaSetupRequiredDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "closed";
          };
}

export interface MfaSetupRequiredDialogOptions {
    props?: {
        onActivate?: () => void;
    };
}
