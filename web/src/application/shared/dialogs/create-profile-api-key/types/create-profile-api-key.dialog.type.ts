export interface CreateProfileApiKeyDialogState {
    state:
        | {
              mode: "open";
          }
        | {
              mode: "closed";
          };
}

export interface CreateProfileApiKeyDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
