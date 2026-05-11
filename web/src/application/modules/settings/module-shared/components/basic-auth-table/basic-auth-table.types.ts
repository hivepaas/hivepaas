export type BasicAuthTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
