export interface RunNowTaskCreatedDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              env: string;
              appId: string;
              scheduledJobId: string;
              taskId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              env: null;
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
