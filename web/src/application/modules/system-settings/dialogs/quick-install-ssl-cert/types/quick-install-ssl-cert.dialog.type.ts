export interface QuickInstallSslCertCreated {
    id: string;
    name: string;
}

export interface QuickInstallSslCertDialogState {
    state:
        | {
              mode: "open";
              domain: string;
          }
        | {
              mode: "closed";
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
