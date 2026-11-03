export interface UpdateAppScheduledJobStatusDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
              scheduledJobId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
              scheduledJobId: null;
          };
}

export interface UpdateAppScheduledJobStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
