export interface RunNowTaskCreatedDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              appId: string;
              scheduledJobId: string;
              taskId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
              scheduledJobId: null;
              taskId: null;
          };
}

export interface RunNowTaskCreatedDialogOptions {
    props?: {
        onClose?: () => void;
    };
}
