export interface CreateOrEditAppScheduledJobDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
          }
        | {
              mode: "edit";
              projectId: string;
              appId: string;
              scheduledJobId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
          };
}

export interface CreateOrEditAppScheduledJobDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
