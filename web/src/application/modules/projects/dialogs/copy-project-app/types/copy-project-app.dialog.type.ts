export interface CopyProjectAppDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
          };
}

export interface CopyProjectAppDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
