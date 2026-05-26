export interface ProjectUserAccessesDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              projectName: string;
          }
        | {
              mode: "closed";
              projectId: null;
              projectName: null;
          };
}

export interface ProjectUserAccessesDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
