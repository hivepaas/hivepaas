export const AppDangerAction = {
    Disable: "disable",
    ReEnable: "re-enable",
    Delete: "delete",
} as const;

export type AppDangerAction = (typeof AppDangerAction)[keyof typeof AppDangerAction];

/** A managed database's cluster a replica was promoted over it in: its members, the primary among them. */
export interface ConfirmAppDangerActionCluster {
    members: { key: string; role: string }[];
}

export interface ConfirmAppDangerActionTarget {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
    updateVer: number;
    /** Set when the app holds a cluster: it is deleted only with all of it. */
    cluster?: ConfirmAppDangerActionCluster;
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
