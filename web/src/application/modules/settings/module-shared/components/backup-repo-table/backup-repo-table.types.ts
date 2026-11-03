export type BackupRepoTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
          env?: string;
      };
