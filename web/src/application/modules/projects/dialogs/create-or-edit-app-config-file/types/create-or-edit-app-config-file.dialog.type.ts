import type { AppConfigFile } from "~/projects/domain";

export interface CreateOrEditAppConfigFileDialogState {
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
              configFile: AppConfigFile;
          }
        | {
              mode: "closed";
              projectId: null;
              appId: null;
          };
}

export interface CreateOrEditAppConfigFileDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
