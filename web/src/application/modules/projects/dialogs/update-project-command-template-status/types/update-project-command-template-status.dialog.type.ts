export interface UpdateProjectCommandTemplateStatusDialogState {
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

export interface UpdateProjectCommandTemplateStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
    };
}
