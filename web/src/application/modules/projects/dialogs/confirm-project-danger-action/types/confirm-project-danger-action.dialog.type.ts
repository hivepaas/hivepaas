export const ProjectDangerAction = {
    Disable: "disable",
    ReEnable: "re-enable",
    Delete: "delete",
} as const;

export type ProjectDangerAction = (typeof ProjectDangerAction)[keyof typeof ProjectDangerAction];

export interface ConfirmProjectDangerActionTarget {
    projectId: string;
    projectName: string;
    updateVer: number;
}

export interface ConfirmProjectDangerActionDialogState {
    state:
        | {
              mode: "open";
              action: ProjectDangerAction;
              target: ConfirmProjectDangerActionTarget;
          }
        | {
              mode: "closed";
              action: null;
              target: null;
          };
}

export interface ConfirmProjectDangerActionDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: (action: ProjectDangerAction) => void;
        onError?: (error: Error) => void;
    };
}
