export interface QuickInstallSslCertCreated {
    id: string;
    name: string;
}

export interface QuickInstallSslCertDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              domain: string;
          }
        | {
              mode: "closed";
              projectId: null;
              domain: null;
          };
}

export interface QuickInstallSslCertDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: (created: QuickInstallSslCertCreated) => void;
        onError?: (error: Error) => void;
    };
}
