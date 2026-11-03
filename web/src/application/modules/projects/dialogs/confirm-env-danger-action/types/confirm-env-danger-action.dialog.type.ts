export const EnvDangerAction = {
    Disable: "disable",
    ReEnable: "re-enable",
    Delete: "delete",
} as const;

export type EnvDangerAction = (typeof EnvDangerAction)[keyof typeof EnvDangerAction];

export interface ConfirmEnvDangerActionTarget {
    projectId: string;
    envName: string;
    updateVer?: number;
}

export interface ConfirmEnvDangerActionDialogState {
    state:
        | {
              mode: "open";
              action: EnvDangerAction;
              target: ConfirmEnvDangerActionTarget;
          }
        | {
              mode: "closed";
              action: null;
              target: null;
          };
}

export interface ConfirmEnvDangerActionDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: (action: EnvDangerAction) => void;
        onError?: (error: Error) => void;
    };
}
