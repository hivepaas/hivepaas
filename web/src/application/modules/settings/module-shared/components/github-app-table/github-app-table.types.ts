export type GithubAppTableScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };
