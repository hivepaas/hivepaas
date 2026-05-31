export interface RunNowTaskCreatedDialogState {
    state:
        | {
              mode: "open";
              taskId: string;
          }
        | {
              mode: "closed";
              taskId: null;
          };
}

export interface RunNowTaskCreatedDialogOptions {
    props?: {
        onClose?: () => void;
    };
}
