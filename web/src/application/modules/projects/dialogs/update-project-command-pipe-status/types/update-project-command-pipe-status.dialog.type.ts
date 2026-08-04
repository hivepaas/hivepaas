export interface UpdateProjectCommandPipeStatusDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              id: string;
          }
        | {
              mode: "closed";
          };
}

export interface UpdateProjectCommandPipeStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
    };
}
