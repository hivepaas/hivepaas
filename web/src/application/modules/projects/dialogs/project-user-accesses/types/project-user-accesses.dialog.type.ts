import type { ProjectEnvEntity } from "~/projects/domain";

export interface ProjectUserAccessesDialogState {
    state:
        | {
              mode: "open";
              projectId: string;
              projectName: string;
              envs: ProjectEnvEntity[];
          }
        | {
              mode: "closed";
              projectId: null;
              projectName: null;
              envs: null;
          };
}

export interface ProjectUserAccessesDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
