export interface CreateFunctionDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
          }
        | {
              mode: "closed";
              projectId: null;
          };
}

export interface CreateFunctionDialogOptions {
    props?: {
        initialEnv?: string;
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
