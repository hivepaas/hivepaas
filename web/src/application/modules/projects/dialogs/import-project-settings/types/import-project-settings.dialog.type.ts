import type { ProjectSettingsImportKind } from "~/projects/data/commands";

export interface ImportProjectSettingsDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              settingKind: ProjectSettingsImportKind;
          }
        | {
              mode: "closed";
              projectId: null;
              settingKind: null;
          };
}

export interface ImportProjectSettingsDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
