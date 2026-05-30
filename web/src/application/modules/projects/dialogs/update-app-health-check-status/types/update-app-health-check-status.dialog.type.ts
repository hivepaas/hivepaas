export interface UpdateAppHealthCheckStatusDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
              healthCheckId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
              healthCheckId: null;
          };
}

export interface UpdateAppHealthCheckStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
