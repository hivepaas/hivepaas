export interface CreateProjectAppDialogState {
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

export interface CreateProjectAppDialogOptions {
    props?: {
        initialEnv?: string;
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
