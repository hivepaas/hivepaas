export interface ExportContainerFilesDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              env: string;
              appId: string;
              nodeId: string;
              containerId: string;
          }
        | {
              mode: "closed";
              projectId: null;
              env: null;
              appId: null;
              nodeId: null;
              containerId: null;
          };
}

export interface ExportContainerFilesDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
