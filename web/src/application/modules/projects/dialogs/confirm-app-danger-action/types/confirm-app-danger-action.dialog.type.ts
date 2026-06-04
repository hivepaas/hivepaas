export const AppDangerAction = {
    Disable: "disable",
    ReEnable: "re-enable",
    Delete: "delete",
} as const;

export type AppDangerAction = (typeof AppDangerAction)[keyof typeof AppDangerAction];

export interface ConfirmAppDangerActionTarget {
    projectId: string;
    appId: string;
    appName: string;
    updateVer: number;
}

export interface ConfirmAppDangerActionDialogState {
    state:
        | {
              mode: "open";
              action: AppDangerAction;
              target: ConfirmAppDangerActionTarget;
          }
        | {
              mode: "closed";
              action: null;
              target: null;
          };
}

export interface ConfirmAppDangerActionDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: (action: AppDangerAction) => void;
        onError?: (error: Error) => void;
    };
}
