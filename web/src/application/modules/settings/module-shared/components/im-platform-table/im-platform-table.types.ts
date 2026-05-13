export type ImPlatformTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
